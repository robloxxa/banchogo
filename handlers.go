package banchogo

var ignoredCodes = []string{
	"312",
	"333",
	"366",
	"372",
	"375",
	"376",
}

var ircHandlers = map[string]func(*BanchoClient, []string){
	"001": handleWelcomeCommand,
}

func handleWelcomeCommand(b *BanchoClient, splits []string) {
	b.setConnectState(Connected)
	//b.emitEvent("connect")
}
