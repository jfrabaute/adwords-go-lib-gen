// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"bytes"
	"go/format"
	"log"
	"os"
	"runtime"

	"github.com/PuerkitoBio/goquery"
	flags "github.com/jessevdk/go-flags"
	gen "github.com/jfrabaute/gowsdl/generator"
)

const (
	version = "v0.0.1"

	adwordsDocURL = "https://developers.google.com/adwords/api/docs/reference/v201409/"
)

var opts struct {
	Version   bool   `short:"v" long:"version" description:"Shows gowsdl version"`
	Package   string `short:"p" long:"package" description:"Package under which code will be generated" default:"myservice"`
	IgnoreTls bool   `short:"i" long:"ignore-tls" description:"Ignores invalid TLS certificates. It is not recomended for production. Use at your own risk" default:"false"`
}

func init() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	log.SetPrefix("🍀  ")
}

type serviceStruct struct {
	name    string
	wsdlURL string
}

func main() {
	/*args*/ _, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		log.Println(version)
		os.Exit(0)
	}

	pkg := "./" + opts.Package
	err = os.Mkdir(pkg, 0744)

	if perr, ok := err.(*os.PathError); ok && os.IsExist(perr.Err) {
		log.Printf("Package directory %s already exist, skipping creation\n", pkg)
	} else {
		if err != nil {
			log.Fatalln(err)
		}
	}

	services, err := getServiceList()
	if err != nil {
		log.Fatalln(err)
	}

	ok := true
	for _, service := range services {
		err = processWsdl(service, opts.Package+"/"+service.name+".go")
		if err != nil {
			log.Println(err)
			ok = false
		}
	}

	if !ok {
		log.Fatal("At least there is one error 💩")
	}
	log.Println("Done 💩")
}

func getServiceList() ([]*serviceStruct, error) {

	list := []string{
		"AdGroupAdService",
		"AdGroupBidModifierService",
		"AdGroupCriterionService",
		"AdGroupFeedService",
		"AdGroupService",
		"AdParamService",
		"AdwordsUserListService",
		"BiddingStrategyService",
		"BudgetOrderService",
		"BudgetService",
		"CampaignAdExtensionService",
		"CampaignCriterionService",
		"CampaignFeedService",
		"CampaignService",
		"CampaignSharedSetService",
		"ConstantDataService",
		"ConversionTrackerService",
		"CustomerFeedService",
		"CustomerService",
		"CustomerSyncService",
		"DataService",
		"ExperimentService",
		"FeedItemService",
		"FeedMappingService",
		"FeedService",
		"GeoLocationService",
		"LabelService",
		"LocationCriterionService",
		"ManagedCustomerService",
		"MediaService",
		"MutateJobService",
		"OfflineConversionFeedService",
		"ReportDefinitionService",
		"SharedCriterionService",
		"SharedSetService",
		"TargetingIdeaService",
		"TrafficEstimatorService",
	}

	var services []*serviceStruct
	for _, name := range list {
		log.Println("Detecting wsdl url for service " + name)
		doc, err := goquery.NewDocument(adwordsDocURL + name)
		if err != nil {
			log.Fatal(err)
		}

		var url string
		doc.Find("dl dd code a").Each(func(i int, s *goquery.Selection) {
			url = s.Text()
		})
		if url == "" {
			log.Fatalln("Unable to find wsdl url for service " + name)
		}
		services = append(services, &serviceStruct{name: name, wsdlURL: url})
	}

	return services, nil
}

func processWsdl(service *serviceStruct, outputFile string) error {
	log.Println("Processing service " + service.name)

	gowsdl, err := gen.NewGoWsdl(service.wsdlURL, opts.Package, opts.IgnoreTls)
	if err != nil {
		return err
	}

	gocode, err := gowsdl.Start()
	if err != nil {
		return err
	}

	fd, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer fd.Close()

	data := new(bytes.Buffer)
	data.Write(gocode["header"])
	data.Write(gocode["types"])
	data.Write(gocode["operations"])

	source, err := format.Source(data.Bytes())
	if err != nil {
		fd.Write(data.Bytes())
		return err
	}

	fd.Write(source)

	return nil
}
