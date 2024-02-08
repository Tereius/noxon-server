package noxon

type NoxonServerSettings struct {
	PresetsModel        PresetModel
	StationsModel       StationsModel
	Whitelist           []string
	Blacklist           []string
	LoginEndpoints      []string
	SearchEndpoints     []string
	GetPresetsEndpoints []string
	AddPresetsEndpoints []string
}

func NewDefaultNoxonServerSettings() NoxonServerSettings {

	return NoxonServerSettings{
		PresetsModel:        NewMemPresetsModel(),
		StationsModel:       NewNullStationsModel(),
		Whitelist:           []string{},
		Blacklist:           []string{},
		LoginEndpoints:      []string{},
		SearchEndpoints:     []string{},
		GetPresetsEndpoints: []string{},
		AddPresetsEndpoints: []string{},
	}
}

func (s NoxonServerSettings) WithPresetsModel(model PresetModel) NoxonServerSettings {

	s.PresetsModel = model
	return s
}

func (s NoxonServerSettings) WithStationsModel(model StationsModel) NoxonServerSettings {

	s.StationsModel = model
	return s
}

func (s NoxonServerSettings) WithBlacklist(list []string) NoxonServerSettings {

	s.Blacklist = list
	return s
}

func (s NoxonServerSettings) WithWhitelist(list []string) NoxonServerSettings {

	s.Whitelist = list
	return s
}

func (s NoxonServerSettings) WithLoginEndpoints(list []string) NoxonServerSettings {

	s.LoginEndpoints = list
	return s
}

func (s NoxonServerSettings) WithSearchEndpoints(list []string) NoxonServerSettings {

	s.SearchEndpoints = list
	return s
}

func (s NoxonServerSettings) WithGetPresetsEndpoints(list []string) NoxonServerSettings {

	s.GetPresetsEndpoints = list
	return s
}

func (s NoxonServerSettings) WithAddPresetsEndpoints(list []string) NoxonServerSettings {

	s.AddPresetsEndpoints = list
	return s
}
