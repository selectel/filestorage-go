/*
Package shareaccess provides access rule operations for the Selectel File Storage API.

The service supports only IP rules, so Allow always sends the ip access type. Each operation sends one
request and returns without waiting for the rule to reach the active state.
*/
package shareaccess
