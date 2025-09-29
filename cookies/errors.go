package cookies

type cookieError string

func (e cookieError) Error() string { return string(e) }
