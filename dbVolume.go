package vlc

type DbVolume struct {
	Left  int
	Right int
}

var dbChan = make(chan DbVolume, 16)

func (p *Player) GetDbVolume() DbVolume {
	select {
	case v, ok := <-dbChan:
		if !ok {
			return DbVolume{}
		}
		return v
	default:
	}
	return DbVolume{}
}
