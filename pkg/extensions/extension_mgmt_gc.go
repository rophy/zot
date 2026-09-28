//go:build mgmt

package extensions

import (
	"encoding/json"
	"errors"
	"net/http"

	zerr "zotregistry.dev/zot/v2/errors"
	"zotregistry.dev/zot/v2/pkg/api/constants"
	zcommon "zotregistry.dev/zot/v2/pkg/common"
	"zotregistry.dev/zot/v2/pkg/log"
	zreg "zotregistry.dev/zot/v2/pkg/regexp"
	reqCtx "zotregistry.dev/zot/v2/pkg/requestcontext"
	"zotregistry.dev/zot/v2/pkg/storage/gc"
)

type GCHandler struct {
	GCOnDemand func(store string) (*gc.OnDemand, bool)
	Log        log.Logger
}

// HandleGC godoc
// @Summary Run garbage collection on demand, or get its status
// @Description POST starts garbage collection of a store, or of a single repository in it.
// @Description GET returns the status of the current or last run, including what it deleted. Admin only.
// @Description A store sweep is paced like the periodic sweep, and may start outside gcTimeWindow.
// @Router  /v2/_zot/ext/mgmt/gc [post]
// @Router  /v2/_zot/ext/mgmt/gc [get]
// @Produce json
// @Param   store   query     string   true    "store route: / for the default store, otherwise a subPaths key"
// @Param   repo    query     string   false   "repository name"
// @Success 200 {object}   gc.RunStatus
// @Success 202 {string}   string   "accepted"
// @Failure 400 {string}   string   "bad request"
// @Failure 403 {string}   string   "forbidden"
// @Failure 404 {string}   string   "store or repository not found"
// @Failure 409 {string}   string   "gc disabled for the store, or already running"
// @Failure 500 {string}   string   "internal server error"
// @Failure 503 {string}   string   "gc could not be scheduled, retry later"
func (h *GCHandler) HandleGC(response http.ResponseWriter, request *http.Request) {
	userAc, err := reqCtx.UserAcFromContext(request.Context())
	if err != nil || userAc == nil || !userAc.IsAdmin() {
		response.WriteHeader(http.StatusForbidden)

		return
	}

	store := request.URL.Query().Get("store")
	if store == "" {
		response.WriteHeader(http.StatusBadRequest)

		return
	}

	onDemand, storeExists := h.GCOnDemand(store)
	if !storeExists {
		response.WriteHeader(http.StatusNotFound)

		return
	}

	if onDemand == nil {
		// GC is disabled for this store
		response.WriteHeader(http.StatusConflict)

		return
	}

	repo := request.URL.Query().Get("repo")
	if repo != "" && !zreg.FullNameRegexp.MatchString(repo) {
		response.WriteHeader(http.StatusBadRequest)

		return
	}

	if request.Method == http.MethodPost {
		h.startGC(response, onDemand, store, repo)

		return
	}

	status := onDemand.Status()

	if repo != "" {
		var found bool

		status, found = onDemand.RepoStatus(repo)
		if !found {
			response.WriteHeader(http.StatusNotFound)

			return
		}
	}

	body, err := json.Marshal(status)
	if err != nil {
		h.Log.Error().Err(err).Str("component", "mgmt").Msg("failed to marshal gc status")
		response.WriteHeader(http.StatusInternalServerError)

		return
	}

	zcommon.WriteData(response, http.StatusOK, constants.DefaultMediaType, body)
}

func (h *GCHandler) startGC(response http.ResponseWriter, onDemand *gc.OnDemand, store, repo string) {
	var err error

	if repo == "" {
		err = onDemand.SweepNow()
	} else {
		err = onDemand.CleanRepoNow(repo)
	}

	switch {
	case err == nil:
		h.Log.Info().Str("component", "mgmt").Str("store", store).Str(constants.RepositoryLogKey, repo).
			Msg("gc requested")
		response.WriteHeader(http.StatusAccepted)
	case errors.Is(err, zerr.ErrInvalidRepositoryName):
		response.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, zerr.ErrRepoNotFound):
		response.WriteHeader(http.StatusNotFound)
	case errors.Is(err, zerr.ErrGCAlreadyRunning):
		response.WriteHeader(http.StatusConflict)
	case errors.Is(err, zerr.ErrGCNotScheduled):
		h.Log.Warn().Err(err).Str("component", "mgmt").Str("store", store).Str(constants.RepositoryLogKey, repo).
			Msg("failed to schedule gc")
		response.WriteHeader(http.StatusServiceUnavailable)
	default:
		h.Log.Error().Err(err).Str("component", "mgmt").Str("store", store).Str(constants.RepositoryLogKey, repo).
			Msg("failed to start gc")
		response.WriteHeader(http.StatusInternalServerError)
	}
}
