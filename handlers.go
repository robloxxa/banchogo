package main

var ircCommands = map[string]func(*BanchoClient, string, []string){
	"001": handleWelcomeCommand,
}

func handleWelcomeCommand(b *BanchoClient, command string, splits []string) {

}
