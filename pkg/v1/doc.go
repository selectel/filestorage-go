/*
Package v1 provides the client and common types for the Selectel File Storage SDK.
The package version is independent of the Manila API version.

The client uses a Manila v2 (sharev2) endpoint and a project-scoped token supplied by the caller.
Every request carries the Manila API microversion 2.75 of the Selectel fork. The client does not
authenticate, retry requests, or wait for resource state changes. API operations are provided by
resource packages such as github.com/selectel/filestorage-go/pkg/v1/share.
*/
package v1
