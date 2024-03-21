module github.com/hotnops/apeman/analyze

go 1.20

require (
	github.com/specterops/bloodhound/analysis v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/cache v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/crypto v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/cypher v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/dawgs v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/errors v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/graphschema v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/headers v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/log v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/mediatypes v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/slices v0.0.0-20240109195535-c358cb6e4aa5
	github.com/specterops/bloodhound/src v0.0.0-20240109195535-c358cb6e4aa5
)

require github.com/go-pkgz/expirable-cache v1.0.0 // indirect

replace github.com/specterops/bloodhound/dawgs => github.com/specterops/bloodhound/packages/go/dawgs v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/cypher => github.com/specterops/bloodhound/packages/go/cypher v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/log => github.com/specterops/bloodhound/packages/go/log v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/analysis => github.com/specterops/bloodhound/packages/go/analysis v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/cache => github.com/specterops/bloodhound/packages/go/cache v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/crypto => github.com/specterops/bloodhound/packages/go/crypto v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/errors => github.com/specterops/bloodhound/packages/go/errors v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/graphschema => github.com/specterops/bloodhound/packages/go/graphschema v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/headers => github.com/specterops/bloodhound/packages/go/headers v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/mediatypes => github.com/specterops/bloodhound/packages/go/mediatypes v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/slices => github.com/specterops/bloodhound/packages/go/slices v0.0.0-20240109195535-c358cb6e4aa5

replace github.com/specterops/bloodhound/src => github.com/SpecterOps/BloodHound/cmd/api/src v0.0.0-20240109195535-c358cb6e4aa5
