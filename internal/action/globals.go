package action

var InfoBar *InfoPane

func InitGlobals() {
	InfoBar = NewInfoBar()
}

func GetInfoBar() *InfoPane {
	return InfoBar
}
