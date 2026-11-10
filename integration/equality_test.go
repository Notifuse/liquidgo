package integration

import (
	"encoding/json"
	"testing"
)

// jsonAssigns decodes data the way applications receive it: every number is a float64.
func jsonAssigns(t *testing.T, data string) map[string]interface{} {
	t.Helper()
	var assigns map[string]interface{}
	if err := json.Unmarshal([]byte(data), &assigns); err != nil {
		t.Fatal(err)
	}
	return assigns
}

func TestEqualityWithJSONData(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     string
		expected string
	}{
		{"number equals integer literal", `{% if v == 100 %}Y{% else %}N{% endif %}`, `{"v":100}`, "Y"},
		{"number differs", `{% if v != 100 %}Y{% else %}N{% endif %}`, `{"v":100.5}`, "Y"},
		{"nested number", `{% if order.total == 100 %}Y{% else %}N{% endif %}`, `{"order":{"total":100}}`, "Y"},
		{"two equal lists", `{% if a == b %}Y{% else %}N{% endif %}`, `{"a":[1,"x"],"b":[1,"x"]}`, "Y"},
		{"two different lists", `{% if a == b %}Y{% else %}N{% endif %}`, `{"a":[1],"b":[2]}`, "N"},
		{"two equal maps", `{% if a == b %}Y{% else %}N{% endif %}`, `{"a":{"k":[1]},"b":{"k":[1]}}`, "Y"},
		{"case when integer", `{% case v %}{% when 3 %}three{% else %}other{% endcase %}`, `{"v":3}`, "three"},
		{"case when list of values", `{% case v %}{% when 1, 2 %}low{% else %}other{% endcase %}`, `{"v":2}`, "low"},
		{"case when or", `{% case v %}{% when 1 or 2 %}low{% else %}other{% endcase %}`, `{"v":2}`, "low"},
		{"contains element", `{% if tags contains "vip" %}Y{% else %}N{% endif %}`, `{"tags":["vip","new"]}`, "Y"},
		{"contains no substring", `{% if tags contains "vip" %}Y{% else %}N{% endif %}`, `{"tags":["vip_gold"]}`, "N"},
		{"contains number", `{% if ids contains 2 %}Y{% else %}N{% endif %}`, `{"ids":[1,2]}`, "Y"},
		{"string contains", `{% if s contains "vip" %}Y{% else %}N{% endif %}`, `{"s":"vip_gold"}`, "Y"},
		{"greater than", `{% if v > 99 %}Y{% else %}N{% endif %}`, `{"v":100}`, "Y"},
		{"where with list values", `{{ items | where: "tags", t | map: "name" | join: "," }}`, `{"items":[{"name":"a","tags":["x"]},{"name":"b","tags":["y"]}],"t":["x"]}`, "a"},
		{"reject with list values", `{{ items | reject: "tags", t | map: "name" | join: "," }}`, `{"items":[{"name":"a","tags":["x"]},{"name":"b","tags":["y"]}],"t":["x"]}`, "b"},
		{"find with list values", `{% assign f = items | find: "tags", t %}{{ f.name }}`, `{"items":[{"name":"a","tags":["x"]},{"name":"b","tags":["y"]}],"t":["y"]}`, "b"},
		{"find_index with list values", `{{ items | find_index: "tags", t }}`, `{"items":[{"name":"a","tags":["x"]},{"name":"b","tags":["y"]}],"t":["y"]}`, "1"},
		{"has with list values", `{{ items | has: "tags", t }}`, `{"items":[{"name":"a","tags":["x"]}],"t":["x"]}`, "true"},
		{"where with a number", `{{ items | where: "n", 2 | map: "name" | join: "," }}`, `{"items":[{"name":"a","n":1},{"name":"b","n":2}]}`, "b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, jsonAssigns(t, tt.data))
		})
	}
}

func TestEqualityWithFilterResults(t *testing.T) {
	assertTemplateResult(t, "even", `{% assign x = 10 | modulo: 2 %}{% if x == 0 %}even{% else %}odd{% endif %}`, nil)
	assertTemplateResult(t, "-2-", `{% for i in (1..3) %}{% if forloop.index == 2 %}-{{ i }}-{% endif %}{% endfor %}`, nil)
}
