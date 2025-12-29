package stats

// Simple GeoIP lookup tool using a local MaxMind DB.
// Used to determine router locations from IP addresses in RouterInfo entries.

type GeoIP struct {
	DBPath string
}

func NewGeoIP(dbPath string) (GeoIP, error) {
	return GeoIP{
		DBPath: dbPath,
	}, nil
}

func (g *GeoIP) Lookup(ip string) (string, error) {
	return "Unknown", nil
}
