# Times the benchmark template with Shopify Liquid, as `libcmp bench` does:
# the median of 15 rounds of at least 200 ms of renders of the template,
# parsed once. ruby -I LIQUID/lib bench.rb templates.json
require "json"
require "liquid"

t = JSON.parse(File.read(ARGV[0]))["bench"]
template = Liquid::Template.parse(t["template"])
data = t["data"]
template.render!(data)
rounds = 15.times.map do
  n = 0
  start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
  while (elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start) < 0.2
    template.render!(data)
    n += 1
  end
  elapsed / n
end
printf("ruby\t%.1fµs\n", rounds.sort[7] * 1e6)
