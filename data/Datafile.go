package data

import (
	"errors"
	"sort"
	"time"
)

type Datafile struct {
	General struct {
		Version          int       `json:"version"`
		Reload           int       `json:"reload"`
		Update           string    `json:"update"`
		UpdateTimestamp  time.Time `json:"update_timestamp"`
		ConnectedClients int       `json:"connected_clients"`
		UniqueUsers      int       `json:"unique_users"`
	} `json:"general"`
	Pilots []struct {
		Cid         int     `json:"cid"`
		Name        string  `json:"name"`
		Callsign    string  `json:"callsign"`
		Server      string  `json:"server"`
		PilotRating int     `json:"pilot_rating"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Altitude    int     `json:"altitude"`
		Groundspeed int     `json:"groundspeed"`
		Transponder string  `json:"transponder"`
		Heading     int     `json:"heading"`
		QnhIHg      float64 `json:"qnh_i_hg"`
		QnhMb       int     `json:"qnh_mb"`
		FlightPlan  *struct {
			FlightRules   string `json:"flight_rules"`
			Aircraft      string `json:"aircraft"`
			AircraftFaa   string `json:"aircraft_faa"`
			AircraftShort string `json:"aircraft_short"`
			Departure     string `json:"departure"`
			Arrival       string `json:"arrival"`
			Alternate     string `json:"alternate"`
			CruiseTas     string `json:"cruise_tas"`
			Altitude      string `json:"altitude"`
			Deptime       string `json:"deptime"`
			EnrouteTime   string `json:"enroute_time"`
			FuelTime      string `json:"fuel_time"`
			Remarks       string `json:"remarks"`
			Route         string `json:"route"`
			RevisionId    int    `json:"revision_id"`
		} `json:"flight_plan"`
		LogonTime   time.Time `json:"logon_time"`
		LastUpdated time.Time `json:"last_updated"`
	} `json:"pilots"`
	Controllers []struct {
		Cid         int       `json:"cid"`
		Name        string    `json:"name"`
		Callsign    string    `json:"callsign"`
		Frequency   string    `json:"frequency"`
		Facility    int       `json:"facility"`
		Rating      int       `json:"rating"`
		Server      string    `json:"server"`
		VisualRange int       `json:"visual_range"`
		TextAtis    []string  `json:"text_atis"`
		LastUpdated time.Time `json:"last_updated"`
		LogonTime   time.Time `json:"logon_time"`
	} `json:"controllers"`
	Atis []struct {
		Cid         int       `json:"cid"`
		Name        string    `json:"name"`
		Callsign    string    `json:"callsign"`
		Frequency   string    `json:"frequency"`
		Facility    int       `json:"facility"`
		Rating      int       `json:"rating"`
		Server      string    `json:"server"`
		VisualRange int       `json:"visual_range"`
		AtisCode    *string   `json:"atis_code"`
		TextAtis    []string  `json:"text_atis"`
		LastUpdated time.Time `json:"last_updated"`
		LogonTime   time.Time `json:"logon_time"`
	} `json:"atis"`
	Servers []struct {
		Ident                    string `json:"ident"`
		HostnameOrIp             string `json:"hostname_or_ip"`
		Location                 string `json:"location"`
		Name                     string `json:"name"`
		ClientsConnectionAllowed int    `json:"clients_connection_allowed"`
	} `json:"servers"`
	Prefiles []struct {
		Cid        int    `json:"cid"`
		Name       string `json:"name"`
		Callsign   string `json:"callsign"`
		FlightPlan struct {
			FlightRules   string `json:"flight_rules"`
			Aircraft      string `json:"aircraft"`
			AircraftFaa   string `json:"aircraft_faa"`
			AircraftShort string `json:"aircraft_short"`
			Departure     string `json:"departure"`
			Arrival       string `json:"arrival"`
			Alternate     string `json:"alternate"`
			CruiseTas     string `json:"cruise_tas"`
			Altitude      string `json:"altitude"`
			Deptime       string `json:"deptime"`
			EnrouteTime   string `json:"enroute_time"`
			FuelTime      string `json:"fuel_time"`
			Remarks       string `json:"remarks"`
			Route         string `json:"route"`
			RevisionId    int    `json:"revision_id"`
		} `json:"flight_plan"`
		LastUpdated time.Time `json:"last_updated"`
	} `json:"prefiles"`
	Facilities []struct {
		Id    int    `json:"id"`
		Short string `json:"short"`
		Long  string `json:"long"`
	} `json:"facilities"`
	Ratings []struct {
		Id    int    `json:"id"`
		Short string `json:"short"`
		Long  string `json:"long"`
	} `json:"ratings"`
	PilotRatings []struct {
		Id        int    `json:"id"`
		ShortName string `json:"short_name"`
		LongName  string `json:"long_name"`
	} `json:"pilot_ratings"`
}

type GeneralInfo struct {
	ConnectedClients int
	UniqueUsers      int
}

type Airfields struct {
	Departures []AirfieldStat
	Arrivals   []AirfieldStat
}

type AirfieldStat struct {
	Icao  string
	Count int
}

func (d *Datafile) GetGeneralInfo() GeneralInfo {
	return GeneralInfo{
		ConnectedClients: d.General.ConnectedClients,
		UniqueUsers:      d.General.UniqueUsers,
	}
}

func (d *Datafile) GetConnectionsPerATCRating() map[string]int {
	atc := map[string]int{
		"ADM": 0,
		"SUP": 0,
		"I3":  0,
		"I1":  0,
		"C3":  0,
		"C1":  0,
		"S3":  0,
		"S2":  0,
		"S1":  0,
	}

	for _, controller := range d.Controllers {
		foundRating, err := d.name(controller.Rating)

		if err != nil {
			panic(err)
		}
		atc[foundRating]++
	}

	return atc
}

func (d *Datafile) name(ratingInt int) (string, error) {
	for _, rating := range d.Ratings {
		if ratingInt == rating.Id {
			return rating.Short, nil
		}
	}

	return "", errors.New("could not match to a rating")
}

func (d *Datafile) GetPopularAirfields() Airfields {
	departures := []AirfieldStat{}

	for _, item := range d.Pilots {
		flightplan := item.FlightPlan

		if flightplan == nil {
			continue
		}

		result, i := airfieldInStructure(flightplan.Departure, departures)

		if !result {
			departures = append(departures, AirfieldStat{
				Icao:  flightplan.Departure,
				Count: 1,
			})
			continue
		}

		if result {
			departures[i].Count++
		}
	}

	sort.Slice(departures, func(i, j int) bool { return departures[i].Count > departures[j].Count })

	return Airfields{
		Departures: departures,
		Arrivals:   nil,
	}
}

func airfieldInStructure(icao string, airfields []AirfieldStat) (bool, int) {
	for i, airfield := range airfields {
		if airfield.Icao == icao {
			return true, i
		}
	}
	return false, 0
}
