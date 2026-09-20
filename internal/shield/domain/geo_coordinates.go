package domain

import "strings"

// LatLng holds geographical latitude and longitude coordinates.
type LatLng struct {
	Lat float64
	Lng float64
}

// countryCoordinates maps lowercased country names and ISO alpha-2/alpha-3 codes to centroid coordinates.
var countryCoordinates = map[string]LatLng{
	// North America
	"united states":               {Lat: 37.0902, Lng: -95.7129},
	"united states of america":   {Lat: 37.0902, Lng: -95.7129},
	"usa":                         {Lat: 37.0902, Lng: -95.7129},
	"us":                          {Lat: 37.0902, Lng: -95.7129},
	"canada":                      {Lat: 56.1304, Lng: -106.3468},
	"ca":                          {Lat: 56.1304, Lng: -106.3468},
	"can":                         {Lat: 56.1304, Lng: -106.3468},
	"mexico":                      {Lat: 23.6345, Lng: -102.5528},
	"mx":                          {Lat: 23.6345, Lng: -102.5528},
	"mex":                         {Lat: 23.6345, Lng: -102.5528},

	// Eurasia / Asia
	"russia":                      {Lat: 55.7558, Lng: 37.6173},
	"russian federation":          {Lat: 55.7558, Lng: 37.6173},
	"ru":                          {Lat: 55.7558, Lng: 37.6173},
	"rus":                         {Lat: 55.7558, Lng: 37.6173},
	"china":                       {Lat: 39.9042, Lng: 116.4074},
	"cn":                          {Lat: 39.9042, Lng: 116.4074},
	"chn":                         {Lat: 39.9042, Lng: 116.4074},
	"north korea":                 {Lat: 39.0392, Lng: 125.7625},
	"democratic people's republic of korea": {Lat: 39.0392, Lng: 125.7625},
	"dprk":                        {Lat: 39.0392, Lng: 125.7625},
	"kp":                          {Lat: 39.0392, Lng: 125.7625},
	"prk":                         {Lat: 39.0392, Lng: 125.7625},
	"south korea":                 {Lat: 35.9078, Lng: 127.7669},
	"republic of korea":           {Lat: 35.9078, Lng: 127.7669},
	"kr":                          {Lat: 35.9078, Lng: 127.7669},
	"kor":                         {Lat: 35.9078, Lng: 127.7669},
	"japan":                       {Lat: 36.2048, Lng: 138.2529},
	"jp":                          {Lat: 36.2048, Lng: 138.2529},
	"jpn":                         {Lat: 36.2048, Lng: 138.2529},
	"iran":                        {Lat: 32.4279, Lng: 53.6880},
	"islamic republic of iran":    {Lat: 32.4279, Lng: 53.6880},
	"ir":                          {Lat: 32.4279, Lng: 53.6880},
	"irn":                         {Lat: 32.4279, Lng: 53.6880},
	"india":                       {Lat: 20.5937, Lng: 78.9629},
	"in":                          {Lat: 20.5937, Lng: 78.9629},
	"ind":                         {Lat: 20.5937, Lng: 78.9629},
	"vietnam":                     {Lat: 14.0583, Lng: 108.2772},
	"vn":                          {Lat: 14.0583, Lng: 108.2772},
	"vnm":                         {Lat: 14.0583, Lng: 108.2772},
	"singapore":                   {Lat: 1.3521, Lng: 103.8198},
	"sg":                          {Lat: 1.3521, Lng: 103.8198},
	"sgp":                         {Lat: 1.3521, Lng: 103.8198},
	"taiwan":                      {Lat: 23.6978, Lng: 120.9605},
	"tw":                          {Lat: 23.6978, Lng: 120.9605},
	"twn":                         {Lat: 23.6978, Lng: 120.9605},
	"hong kong":                   {Lat: 22.3193, Lng: 114.1694},
	"hk":                          {Lat: 22.3193, Lng: 114.1694},
	"hkg":                         {Lat: 22.3193, Lng: 114.1694},
	"israel":                      {Lat: 31.0461, Lng: 34.8516},
	"il":                          {Lat: 31.0461, Lng: 34.8516},
	"isr":                         {Lat: 31.0461, Lng: 34.8516},
	"turkey":                      {Lat: 38.9637, Lng: 35.2433},
	"tr":                          {Lat: 38.9637, Lng: 35.2433},
	"tur":                         {Lat: 38.9637, Lng: 35.2433},
	"pakistan":                    {Lat: 30.3753, Lng: 69.3451},
	"pk":                          {Lat: 30.3753, Lng: 69.3451},
	"indonesia":                   {Lat: -0.7893, Lng: 113.9213},
	"id":                          {Lat: -0.7893, Lng: 113.9213},
	"malaysia":                    {Lat: 4.2105, Lng: 101.9758},
	"my":                          {Lat: 4.2105, Lng: 101.9758},
	"philippines":                 {Lat: 12.8797, Lng: 121.7740},
	"ph":                          {Lat: 12.8797, Lng: 121.7740},
	"thailand":                    {Lat: 15.8700, Lng: 100.9925},
	"th":                          {Lat: 15.8700, Lng: 100.9925},
	"united arab emirates":        {Lat: 23.4241, Lng: 53.8478},
	"uae":                         {Lat: 23.4241, Lng: 53.8478},
	"ae":                          {Lat: 23.4241, Lng: 53.8478},
	"saudi arabia":                {Lat: 23.8859, Lng: 45.0792},
	"sa":                          {Lat: 23.8859, Lng: 45.0792},

	// Europe
	"united kingdom":              {Lat: 55.3781, Lng: -3.4360},
	"uk":                          {Lat: 55.3781, Lng: -3.4360},
	"gb":                          {Lat: 55.3781, Lng: -3.4360},
	"gbr":                         {Lat: 55.3781, Lng: -3.4360},
	"germany":                     {Lat: 51.1657, Lng: 10.4515},
	"de":                          {Lat: 51.1657, Lng: 10.4515},
	"deu":                         {Lat: 51.1657, Lng: 10.4515},
	"france":                      {Lat: 46.2276, Lng: 2.2137},
	"fr":                          {Lat: 46.2276, Lng: 2.2137},
	"fra":                         {Lat: 46.2276, Lng: 2.2137},
	"netherlands":                 {Lat: 52.1326, Lng: 5.2913},
	"nl":                          {Lat: 52.1326, Lng: 5.2913},
	"nld":                         {Lat: 52.1326, Lng: 5.2913},
	"ukraine":                     {Lat: 48.3794, Lng: 31.1656},
	"ua":                          {Lat: 48.3794, Lng: 31.1656},
	"ukr":                         {Lat: 48.3794, Lng: 31.1656},
	"belarus":                     {Lat: 53.7098, Lng: 27.9534},
	"by":                          {Lat: 53.7098, Lng: 27.9534},
	"blr":                         {Lat: 53.7098, Lng: 27.9534},
	"poland":                      {Lat: 51.9194, Lng: 19.1451},
	"pl":                          {Lat: 51.9194, Lng: 19.1451},
	"pol":                         {Lat: 51.9194, Lng: 19.1451},
	"romania":                     {Lat: 45.9432, Lng: 24.9668},
	"ro":                          {Lat: 45.9432, Lng: 24.9668},
	"rou":                         {Lat: 45.9432, Lng: 24.9668},
	"switzerland":                 {Lat: 46.8182, Lng: 8.2275},
	"ch":                          {Lat: 46.8182, Lng: 8.2275},
	"che":                         {Lat: 46.8182, Lng: 8.2275},
	"sweden":                      {Lat: 60.1282, Lng: 18.6435},
	"se":                          {Lat: 60.1282, Lng: 18.6435},
	"swe":                         {Lat: 60.1282, Lng: 18.6435},
	"norway":                      {Lat: 60.4720, Lng: 8.4689},
	"no":                          {Lat: 60.4720, Lng: 8.4689},
	"nor":                         {Lat: 60.4720, Lng: 8.4689},
	"ireland":                     {Lat: 53.1424, Lng: -7.6921},
	"ie":                          {Lat: 53.1424, Lng: -7.6921},
	"irl":                         {Lat: 53.1424, Lng: -7.6921},
	"spain":                       {Lat: 40.4637, Lng: -3.7492},
	"es":                          {Lat: 40.4637, Lng: -3.7492},
	"esp":                         {Lat: 40.4637, Lng: -3.7492},
	"italy":                       {Lat: 41.8719, Lng: 12.5674},
	"it":                          {Lat: 41.8719, Lng: 12.5674},
	"ita":                         {Lat: 41.8719, Lng: 12.5674},

	// Africa
	"nigeria":                     {Lat: 9.0820, Lng: 8.6753},
	"ng":                          {Lat: 9.0820, Lng: 8.6753},
	"nga":                         {Lat: 9.0820, Lng: 8.6753},
	"south africa":                {Lat: -30.5595, Lng: 22.9375},
	"za":                          {Lat: -30.5595, Lng: 22.9375},
	"zaf":                         {Lat: -30.5595, Lng: 22.9375},
	"egypt":                       {Lat: 26.8206, Lng: 30.8025},
	"eg":                          {Lat: 26.8206, Lng: 30.8025},
	"egy":                         {Lat: 26.8206, Lng: 30.8025},
	"kenya":                       {Lat: -0.0236, Lng: 37.9062},
	"ke":                          {Lat: -0.0236, Lng: 37.9062},
	"ken":                         {Lat: -0.0236, Lng: 37.9062},

	// South America / Oceania
	"brazil":                      {Lat: -14.2350, Lng: -51.9253},
	"br":                          {Lat: -14.2350, Lng: -51.9253},
	"bra":                         {Lat: -14.2350, Lng: -51.9253},
	"argentina":                   {Lat: -38.4161, Lng: -63.6167},
	"ar":                          {Lat: -38.4161, Lng: -63.6167},
	"arg":                         {Lat: -38.4161, Lng: -63.6167},
	"colombia":                    {Lat: 4.5709, Lng: -74.2973},
	"co":                          {Lat: 4.5709, Lng: -74.2973},
	"col":                         {Lat: 4.5709, Lng: -74.2973},
	"chile":                       {Lat: -35.6751, Lng: -71.5430},
	"cl":                          {Lat: -35.6751, Lng: -71.5430},
	"chl":                         {Lat: -35.6751, Lng: -71.5430},
	"australia":                   {Lat: -25.2744, Lng: 133.7751},
	"au":                          {Lat: -25.2744, Lng: 133.7751},
	"aus":                         {Lat: -25.2744, Lng: 133.7751},
	"new zealand":                 {Lat: -40.9006, Lng: 174.8860},
	"nz":                          {Lat: -40.9006, Lng: 174.8860},
	"nzl":                         {Lat: -40.9006, Lng: 174.8860},
}

// ResolveCountryCoordinates returns the centroid coordinates for a given country name or code.
// If the country is unrecognized, a fallback coordinate is returned.
func ResolveCountryCoordinates(country string) (lat float64, lng float64) {
	clean := strings.ToLower(strings.TrimSpace(country))
	if coords, ok := countryCoordinates[clean]; ok {
		return coords.Lat, coords.Lng
	}

	// Try substring matching for full names (skip short codes)
	if len(clean) >= 4 {
		for name, coords := range countryCoordinates {
			if len(name) >= 4 && (strings.Contains(name, clean) || strings.Contains(clean, name)) {
				return coords.Lat, coords.Lng
			}
		}
	}

	return 20.0, 0.0
}
