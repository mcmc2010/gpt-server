package server

import database_level "mcmcx.com/gpt-server/database/level"

func LDBReleaseAll() {
	database_level.ReleaseAll()
}
