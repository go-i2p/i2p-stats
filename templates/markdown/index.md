Exploratory Build Stats Log
---------------------------
{{range .Stats}}
### Stats for: {{.CollectedDate}}

#### Build Statistics

 - Exploratory Build Success Percentage: {{.ExploratoryBuildSucceededPercent}}
 - Exploratory Build Rejection Percentage: {{.ExploratoryBuildRejectedPercent}}
 - Exploratory Build Expired Percentage: {{.ExploratoryBuildExpiredPercent}}
 - Exploratory Build Success: {{.ExploratoryBuildSucceeded}}
 - Exploratory Build Reject: {{.ExploratoryBuildRejected}}
 - Exploratory Build Expired: {{.ExploratoryBuildExpired}}
{{if .DHTEnabled}}

#### DHT Network Statistics
{{if gt .TotalRouters 0}}
 - Total Routers: {{.TotalRouters}}
 - IPv4 Routers: {{.IPv4Routers}}
 - IPv6 Routers: {{.IPv6Routers}}
 - Floodfill Routers: {{.FloodfillRouters}}
 - Reachable Routers: {{.ReachableRouters}}
 - NTCP2 Routers: {{.NTCP2Routers}}
 - SSU2 Routers: {{.SSU2Routers}}
{{else}}
 - DHT collection enabled but no routers found in netDb
{{end}}
{{end}}
{{end}}
