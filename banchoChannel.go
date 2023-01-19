package banchogo

type BanchoChannel struct {
	Name    string
	Topic   string
	Joined  bool
	Members map[string]BanchoChannelMember
}
