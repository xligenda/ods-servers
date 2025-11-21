package structs

type User struct {
	ID DiscordID
	// guilds where this user has a membership
	Servers map[ServerTag][]RoleName
}
