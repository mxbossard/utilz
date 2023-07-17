package serializ

var (
	jsonEmpty0 = ``
	jsonEmpty1 = `{}`
	json0      = `{"key": "value"}`
	json1c     = `{"K1": "pif", "k2": "paf"}`
	json1c2    = `{"k1": "puf", "K2": "pef"}`
	json1      = `{"s1": "foo", "a1": ["bar", "baz"], "m1": ` + json1c + `}`
	json2      = `{"A2": [` + json1 + `], "A3": [` + json1c + `, ` + json1c2 + `]}`
	json3      = `{"A4": [{"A4": [{"A4": [{"A4": []}]}]}]}`
)

func TestTransform_JsonAdd(t *testing.T) {
	op := OpAdd("foo", 4)

	var in map[string]any
	out, err := op.Transform(in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "foo")
	assert.Equals(t, out["foo"], 4)
}
