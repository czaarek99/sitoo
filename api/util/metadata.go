package util

/*
Currently the metadata only has a RequestID which is handy
for tracking logs.

In a production service we would probably have more information
here as for example the user that is actually making
the request.
*/
type Metadata struct {
	RequestID uint32
}
