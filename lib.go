package main

func interface2StrSlice(i []interface{}) []string {
	s := []string{}
	for _, v := range i {
		s = append(s, v.(string))
	}
	return s
}

func mapInterface2MapStr(m map[string]interface{}) map[string]string {
	s := map[string]string{}
	for k, v := range m {
		s[k] = v.(string)
	}
	return s
}
