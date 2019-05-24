package fsm

type IState interface {
	Enter(stateFrom *IState) error
	Update() error
	Exit(stateTo *IState) error
}

type State struct {
	name   string
	active bool
}

func MakeState(n string) State {
	return State{
		name: n,
	}
}

func NewState(n string) *State {
	s := MakeState(n)
	return &s
}

type IMachine interface {
	Add(s *IState)
	Remove(s *IState)
	Enter(stateFrom *IState) error
	Update() error
	Exit(stateTo *IState) error
}

type Machine struct {
	currentStateIndex  int
	nextStateIndex     int
	previousStateIndex int
	numStates          int
	active             bool
	states             []*IState
}

func MakeMachine() Machine {
	return Machine{
		states: make([]*IState, 0, 10),
	}
}

func NewMachine() *Machine {
	m := MakeMachine()
	return &m
}

func (m *Machine) Add(s *IState) {
	//if m.numStates >= len(m.states) {
	//newStatesLen := m.numStates+2
	//m.states = make([]*State, newStatesLen, 10)
	m.states = append(m.states, s)
	//}

	//m.states[m.numStates] = s
	m.numStates++
}

func (m *Machine) Remove(s *IState) {
	currentStateIndex := 0
	for nextState := m.states[currentStateIndex]; nil != nextState; currentStateIndex++ {
		if nextState == s {
			nextState = nil
		}
	}
}

func (m *Machine) Enter(stateFrom *IState) error {
	var s IState = *m.states[m.currentStateIndex]
	err := s.Enter(stateFrom)
	return err
}

func (m *Machine) Update() error {
	return nil
}

func (m *Machine) Exit(stateTo *IState) error {
	var s IState = *m.states[m.currentStateIndex]
	err := s.Exit(stateTo)
	return err
}
