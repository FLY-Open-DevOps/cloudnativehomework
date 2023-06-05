package module1

func ModifySlice(slice []string) {
	for i, str := range slice {
		switch str {
		case "stupid":
			slice[i] = "smart"
		case "weak":
			slice[i] = "strong"
		default:
		}
	}
}
