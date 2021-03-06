package spaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/simonleung8/cli-stack-changer/orgs"
)

type Spaces interface {
	GetSpaceGuid(flags.FlagContext) (string, error)
}

type SpacesModel struct {
	NextUrl   string       `json:"next_url,omitempty"`
	Resources []SpaceModel `json:"resources"`
}

type MetadataModel struct {
	Guid string `json:"guid"`
}

type EntityModel struct {
	Name             string `json:"name"`
	OrganizationGuid string `json:"organization_guid"`
}

type SpaceModel struct {
	Metadata MetadataModel `json:"metadata"`
	Entity   EntityModel   `json:"entity"`
}

type spaces struct {
	cliCon plugin.CliConnection
}

func NewSpaces(cliConnection plugin.CliConnection) Spaces {
	return &spaces{
		cliCon: cliConnection,
	}
}

func (s *spaces) GetSpaceGuid(fc flags.FlagContext) (string, error) {
	var spaceExists bool
	nextUrl := "/v2/spaces"

	for nextUrl != "" {
		output, err := s.cliCon.CliCommandWithoutTerminalOutput("curl", nextUrl)
		if err != nil {
			return "", err
		}

		model := SpacesModel{}
		err = json.Unmarshal([]byte(output[0]), &model)
		if err != nil {
			return "", err
		}

		for _, space := range model.Resources {
			if strings.ToLower(space.Entity.Name) == strings.ToLower(fc.String("s")) {
				spaceExists = true
				orgObj := orgs.NewOrgs(s.cliCon)
				org, err := orgObj.GetOrg(fc.String("o"))
				if err != nil {
					return "", err
				}

				if org.Metadata.Guid == "" {
					return "", errors.New(fmt.Sprintf("Organization '%s' does not exist", fc.String("o")))
				}

				if space.Entity.OrganizationGuid == org.Metadata.Guid {
					return space.Metadata.Guid, nil
				}
			}
		}

		if model.NextUrl != "" {
			nextUrl = model.NextUrl
		} else {
			nextUrl = ""
		}
	}

	if spaceExists {
		return "", errors.New(fmt.Sprintf("Space '%s' does not belong to Organization '%s'", fc.String("s"), fc.String("o")))
	}

	return "", errors.New(fmt.Sprintf("Space '%s' does not exist.", fc.String("s")))
}
