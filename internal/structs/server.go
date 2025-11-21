package structs

type Server struct {
	Tag   ServerTag
	Roles map[DiscordID]RoleName
}

type ServerTag = int
type DiscordID = int
type RoleName = string
