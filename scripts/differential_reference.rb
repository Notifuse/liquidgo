# frozen_string_literal: true

# Renders the differential corpus with Shopify Liquid and writes the results
# that integration/differential_test.go compares liquidgo with.
#
# Usage, from the root of the repository, with a checkout of Shopify Liquid
# v5.11.0 (its gems strscan >= 3.1.1 and bigdecimal installed):
#
#   TZ=UTC ruby -I <shopify-liquid>/lib scripts/differential_reference.rb         # write
#   TZ=UTC ruby -I <shopify-liquid>/lib scripts/differential_reference.rb --check # compare
#
# The corpus is integration/testdata/differential/templates.json, written by
# hand, plus generated templates: fragments of tags, filters and delimiters
# joined at random from a fixed seed, so that every run writes the same file.

require "json"
require "liquid"

DIR = File.expand_path("../integration/testdata/differential", __dir__)
EXPECTED = File.join(DIR, "expected.json")
MODES = %w[lax strict].freeze
GENERATED = 400

# The generated templates are built from one or two whole tags or outputs,
# nested at most one level, over the names of data.json; one in ten loses a character, so that
# malformed markup is generated too. Loops are short: the corpus is about
# what a template renders, not about its cost.
VARIABLES = %w[x name arr words n f s posts h empty nothing missing].freeze
FILTERS = [
  "upcase", "downcase", "capitalize", "size", "first", "last", "join: ','", "split: ','", "escape", "strip",
  "append: 'z'", "prepend: 'z'", "plus: 1", "minus: 2.5", "times: 3", "divided_by: 2", "modulo: 2",
  "default: 'd'", "slice: 1, 2", "slice: -2", "truncate: 4", "truncatewords: 1", "sort", "reverse", "uniq",
  "compact", "map: 'title'", "where: 'public'", "replace: 'l', 'L'", "remove: 'o'", "url_encode", "abs",
  "round: 1", "ceil", "floor", "at_least: 2", "at_most: 2", "strip_html", "newline_to_br", "concat: arr",
].freeze
LITERALS = ["'lit'", "\"q\"", "1", "-2", "2.5", "true", "false", "nil", "empty", "blank", "(1..3)"].freeze

class Generator
  def initialize(seed) = @rng = Random.new(seed)

  def pick(list) = list[@rng.rand(list.size)]
  def chance(p) = @rng.rand < p

  def expression
    return pick(LITERALS) if chance(0.3)

    v = pick(VARIABLES)
    v += pick([".size", ".first", ".last", "[0]", "[1]", ".k", "['k']", ".title", "[-1]"]) if chance(0.3)
    v
  end

  def filtered = ([expression] + Array.new(@rng.rand(0..2)) { pick(FILTERS) }).join(" | ")
  def condition = [expression, pick(["==", "!=", "<", ">", ">=", "contains"]), expression].join(" ") +
                  (chance(0.3) ? " #{pick(%w[and or])} #{expression}" : "")

  def body(depth) = Array.new(@rng.rand(1..2)) { node(depth + 1) }.join

  def node(depth)
    kind = depth >= 2 ? pick(%i[text output output assign]) : pick(%i[text output output assign capture if unless case for tablerow cycle increment echo raw comment liquid])
    case kind
    when :text then pick(["a", " ", "<p>", "\n", "é"])
    when :output then "{{#{pick(['', '-'])} #{filtered} #{pick(['', '-'])}}}"
    when :assign then "{% assign #{pick(%w[v w x])} = #{filtered} %}"
    when :capture then "{% capture c %}#{body(depth)}{% endcapture %}{{ c }}"
    when :if
      elsif_part = chance(0.3) ? "{% elsif #{condition} %}#{body(depth)}" : ""
      else_part = chance(0.4) ? "{% else %}#{body(depth)}" : ""
      "{% if #{condition} %}#{body(depth)}#{elsif_part}#{else_part}{% endif %}"
    when :unless then "{% unless #{condition} %}#{body(depth)}{% endunless %}"
    when :case
      whens = Array.new(@rng.rand(1..2)) { "{% when #{expression} %}#{body(depth)}" }.join
      "{% case #{expression} %}#{whens}#{chance(0.5) ? "{% else %}#{body(depth)}" : ''}{% endcase %}"
    when :for
      options = [pick(["", " limit: 2", " offset: 1", " reversed", " offset: 1 limit: 1"])].join
      "{% for i in #{pick(%w[arr words posts (1..3) (3..1) s h nothing])}#{options} %}" \
        "#{pick(['', '{{ i }}', '{{ forloop.index }}'])}#{body(depth)}" \
        "#{chance(0.2) ? '{% else %}e' : ''}{% endfor %}"
    when :tablerow then "{% tablerow i in #{pick(%w[arr words (1..3)])}#{pick(['', ' cols: 2', ' limit: 2'])} %}{{ i }}#{body(depth)}{% endtablerow %}"
    when :cycle then "{% cycle #{pick(["'a', 'b'", "'g': 1, 2", "x, n"])} %}"
    when :increment then "{% #{pick(%w[increment decrement])} #{pick(%w[k n x])} %}"
    when :echo then "{% echo #{filtered} %}"
    when :raw then "{% raw %}{{ #{expression} }}{% endraw %}"
    when :comment then "{% comment %}#{body(depth)}{% endcomment %}"
    when :liquid then "{% liquid\nassign v = #{filtered}\necho v\n%}"
    end
  end

  def template
    t = Array.new(@rng.rand(1..2)) { node(1) }.join
    if chance(0.1) && t.size > 1
      i = @rng.rand(t.size)
      t = t[0...i] + t[(i + 1)..]
    end
    t
  end
end

def generated
  generator = Generator.new(11)
  (1..GENERATED).map { |i| { "id" => format("generated-%03d", i), "template" => generator.template } }
end

def render(template, data, mode)
  parsed = Liquid::Template.parse(template, error_mode: mode.to_sym)
  output = parsed.render(Marshal.load(Marshal.dump(data)))
  { "output" => output.dup.force_encoding("UTF-8").scrub("�") }
rescue Liquid::Error => e
  { "error" => "#{e.class}: #{e.message}" }
rescue StandardError, SystemStackError => e
  { "error" => "ruby #{e.class}" }
end

templates = JSON.parse(File.read(File.join(DIR, "templates.json"), encoding: "UTF-8")) + generated
data = JSON.parse(File.read(File.join(DIR, "data.json"), encoding: "UTF-8"))
results = templates.flat_map do |t|
  MODES.map { |mode| { "id" => t["id"], "mode" => mode, "template" => t["template"] }.merge(render(t["template"], data, mode)) }
end
text = JSON.pretty_generate({ "liquid" => Liquid::VERSION, "results" => results }) + "\n"

if ARGV.include?("--check")
  if File.binread(EXPECTED) != text.b
    warn "#{EXPECTED} is not what Shopify Liquid #{Liquid::VERSION} renders: run this script without --check"
    exit 1
  end
else
  File.binwrite(EXPECTED, text)
end
