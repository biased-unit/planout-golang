FUNC_TEMPLATE = """func (t *TypedMap) get%(n)s(key string) (%(t)s, bool) {
	value, exists := t.get(key)

	if !exists {
		return %(empty)s, false
	}

	if reflect.TypeOf(value).String() != "%(t)s" {
		return %(empty)s, false
	}

	return value.(%(t)s), true

}
"""

TYPES = [
    dict(n='String', t='string', empty='""'),
    dict(n='Bool', t='bool', empty='false'),
    dict(n='Int', t='int', empty='0'),
    dict(n='Int64', t='int64', empty='0'),
    dict(n='Float32', t='float32', empty='0.0'),
    dict(n='Float64', t='float64', empty='0.0'),
    dict(n='Map', t='map[string]interface {}', empty='nil'),
    dict(n='Array', t='[]interface {}', empty='nil')
]

if __name__ == '__main__':
    for t in TYPES:
        print(FUNC_TEMPLATE % t)
