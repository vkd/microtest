package vars

type Map map[string]interface{}

func (m Map) Merge(vs Map) {
	if m == nil {
		m = Map{}
	}
	for k, v := range vs {
		m[k] = v
	}
}

func (m Map) Add(key string, value interface{}) {
	if m == nil {
		m = Map{}
	}
	m[key] = value
}
