/*
Package share provides share operations for the Selectel File Storage API.

Create uses the flat request body of the Selectel Manila fork: the fork creates the service share
network and port together with the share and removes them when the share is deleted. Each operation
sends one request and returns without waiting for a resource state change.
*/
package share
