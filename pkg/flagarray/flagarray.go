package flagarray

type FlagArray []string

func (i *FlagArray) String() string {
	return "FlagArray"
}

func (i *FlagArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *FlagArray) Insert(value string) error {
	*i = append([]string{value}, *i...)
	return nil
}
