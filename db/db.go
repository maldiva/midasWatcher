package db

/// imitation of retrieving values from database

var watchAddresses = []string{"medRDLxsMaBxj1EqxkKtWPnXYX2VEgDEYf", "maWSPAoX4kP3CSsYzeZCZ2auuzP71jZSzB"}

var lvbh = "43fb88f4f18edce012469b5a3047d7a36c127eb884eba1839aae6364e99b2298"

func GetLastVisibleBlockHash() string {
	return lvbh
}

func SetLastVisibleBlockHash(hash string) {
	lvbh = hash
}

func GetWatchAddresses() []string {
	return watchAddresses
}
