package main

import (
	"log"

	"code.google.com/p/dsallings-couch-go"
)

type Doc struct {
	Trigger string
	Target  string
	Url     string
}

func updateDBConf(il interestingList, doc Doc) {
	l, ok := il[doc.Target]
	if !ok {
		l = make([]string, 0, 1)
	}
	il[doc.Target] = append(l, doc.Url)
}

func loadInterestingCouch(dburl string) Subscriber {
	db, err := couch.Connect(dburl)
	maybefatal(err, "Error connecting: %v", err)

	type hookResults struct {
		Rows []struct {
			Doc Doc `json:"doc"`
		}
	}

	results := hookResults{}
	err = db.Query("_design/app/_view/hooks", map[string]interface{}{
		"include_docs": true,
	}, &results)
	maybefatal(err, "Problem loading hooks from DB: %v", err)

	conf := jsonConfig{}
	conf.Actor = make(interestingList)
	conf.Owner = make(interestingList)
	conf.OwnerRepo = make(interestingList)

	for _, doc := range results.Rows {
		switch doc.Doc.Trigger {
		case "ByActor":
			updateDBConf(conf.Actor, doc.Doc)
		case "ByOwner":
			updateDBConf(conf.Owner, doc.Doc)
		case "ByOwnerRepo":
			updateDBConf(conf.OwnerRepo, doc.Doc)
		default:
			log.Fatalf("Unknown trigger: %v", doc.Doc.Trigger)
		}
	}

	return &conf
}
