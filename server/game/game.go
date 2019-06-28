package game
var currentGames [][2]*player

// enter -1 for id to ignore to search with session id
func getGameIndex(id int, sessionID int) (int, int) {
	playerMux.Lock()
	defer playerMux.Unlock()

	for i, val := range currentGames {
		for j, val := range val {
			if id != -1 {
				if val.id == id {
					return i, j
				}
			} else {
				if val.sessionID == sessionID {
					return i, j
				}
			}
		}
	}
	return -1, -1
}

