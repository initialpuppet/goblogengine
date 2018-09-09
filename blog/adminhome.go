package blog

import (
	"context"
	"goblogengine/model"
	"net/http"
	"time"

	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
)

type adminHomeViewModel struct {
	PostCount     int
	DraftCount    int
	VersionCount  int
	AuthorCount   int
	CategoryCount int
	ImageCount    int

	AuditEvents []auditEventViewModel
}

type auditEventViewModel struct {
	When time.Time
	Who  string
	Text string
}

func (a *adminHomeViewModel) addStatistics(s *model.Statistics) {
	a.PostCount = s.PostCount
	a.DraftCount = s.DraftCount
	a.VersionCount = s.VersionCount
	a.AuthorCount = s.AuthorCount
	a.CategoryCount = s.CategoryCount
	a.ImageCount = s.ImageCount
}

func (a *adminHomeViewModel) addAuditEvents(evts []model.Audit) {
	for i := range evts {
		v := auditEventViewModel{
			When: evts[i].When,
			Who:  evts[i].Author.Email,
			Text: evts[i].String(),
		}
		a.AuditEvents = append(a.AuditEvents, v)
	}
}

// AdminHomeGET displays the admin home page.
func AdminHomeGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	s, err := model.GetStatistics(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	evts, err := model.GetAuditTail(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	viewModel := new(adminHomeViewModel)
	viewModel.addStatistics(s)
	viewModel.addAuditEvents(evts)

	v := env.View.New("admin/adminhome")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}
