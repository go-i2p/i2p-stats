package stats

import (
	"log"
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
)

// Simple GeoIP lookup tool using a local MaxMind DB.
// Used to determine router locations from IP addresses in RouterInfo entries.
type GeoIP struct {
	DBPath string
	DB     *geoip2.Reader
}

func NewGeoIP(dbPath string) (GeoIP, error) {
	return GeoIP{
		DBPath: dbPath,
	}, nil
}

func (g *GeoIP) Lookup(ip string) (string, error) {
	parsedIP, err := netip.ParseAddr(ip)
	if err != nil {
		log.Fatal(err)
	}
	city, err := g.City(parsedIP)
	if err != nil {
		return "", err
	}
	country, err := g.Country(parsedIP)
	if err != nil {
		return "", err
	}
	return city + ", " + country, nil
}

func (g *GeoIP) City(parsedIP netip.Addr) (string, error) {
	city, err := g.DB.City(parsedIP)
	if err != nil {
		return "", err
	}
	plainCity := "Unknown City"
	if city != nil {
		plainCity = city.City.Names.English
	}
	return plainCity, nil
}

func (g *GeoIP) Country(parsedIP netip.Addr) (string, error) {
	country, err := g.DB.Country(parsedIP)
	if err != nil {
		return "", err
	}
	plainCountry := "Unknown Country"
	if country != nil {
		plainCountry = country.Country.Names.English
	}
	return plainCountry, nil
}
