package stats

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-i2p/common/router_info"
	"github.com/go-i2p/logger"
)

var log = logger.GetGoI2PLogger()

// In this file we do deeper analysis of the DHT data than we can do with I2PControl alone.
// We collect the RouterInfo entries from our local view of the DHT and analyze them.

type DHT struct {
	RouterInfos []router_info.RouterInfo
	Path        string
}

func NewDHT(dhtDir string) (DHT, error) {
	dht := DHT{
		Path: dhtDir,
	}
	riset, err := dht.routerInfos()
	if err != nil {
		return dht, err
	}
	dht.RouterInfos = riset
	return dht, nil
}

func (db *DHT) routerInfos() (routerInfos []router_info.RouterInfo, err error) {
	r, _ := regexp.Compile("^routerInfo-[A-Za-z0-9-=~]+.dat$")

	files := make(map[string]os.FileInfo)
	walkpath := func(path string, f os.FileInfo, err error) error {
		if r.MatchString(f.Name()) {
			files[path] = f
		}
		return nil
	}

	filepath.Walk(db.Path, walkpath)

	for path := range files {
		riBytes, err := os.ReadFile(path)
		if nil != err {
			log.WithError(err).WithField("path", path).Error("Error reading RouterInfo file")
			continue
		}

		riStruct, remainder, err := router_info.ReadRouterInfo(riBytes)
		if err != nil {
			log.WithError(err).WithField("path", path).Error("RouterInfo Parsing Error")
			log.WithField("path", path).WithField("remainder", remainder).Debug("Leftover Data(for debugging)")
			continue
		} else {
			log.WithField("path", path).Debug("Successfully parsed RouterInfo")
		}

		routerInfos = append(routerInfos, riStruct)
	}

	return routerInfos, err
}

func (db *DHT) CountRouters() int {
	return len(db.RouterInfos)
}

func (db *DHT) CountIPv4Routers() int {
	count := 0
	for _, ri := range db.RouterInfos {
		if ri.HasIPv4() {
			count++
		}
	}
	return count
}

func (db *DHT) CountIPv6Routers() int {
	count := 0
	for _, ri := range db.RouterInfos {
		if ri.HasIPv6() {
			count++
		}
	}
	return count
}

func (db *DHT) CountFloodfills() int {
	count := 0
	for _, ri := range db.RouterInfos {
		if ri.IsFloodfill() {
			count++
		}
	}
	return count
}

func (db *DHT) CountReachableRouters() int {
	count := 0
	for _, ri := range db.RouterInfos {
		if ri.Reachable() {
			count++
		}
	}
	return count
}

func (db *DHT) CountNTCP2Routers() int {
	count := 0
	for _, ri := range db.RouterInfos {
		if ri.SupportsNTCP2() {
			count++
		}
	}
	return count
}

func (db *DHT) CountSSU2Routers() int {
	count := 0
	for _, ri := range db.RouterInfos {
		if ri.SupportsSSU2() {
			count++
		}
	}
	return count
}
