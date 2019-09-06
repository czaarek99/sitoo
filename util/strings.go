package util

func StringsToInterfaces(strings []string) []interface{} {

	interfaces := make([]interface{}, len(strings))

	for i := 0; i < len(strings); i++ {
		interfaces[i] = strings[i]
	}

	return interfaces
}
