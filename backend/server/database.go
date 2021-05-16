package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/bytebase/bytebase"
	"github.com/bytebase/bytebase/api"
	"github.com/google/jsonapi"
	"github.com/labstack/echo/v4"
)

func (s *Server) registerDatabaseRoutes(g *echo.Group) {
	g.POST("/database", func(c echo.Context) error {
		databaseCreate := &api.DatabaseCreate{WorkspaceId: api.DEFAULT_WORKPSACE_ID}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, databaseCreate); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformatted create database request").SetInternal(err)
		}

		databaseCreate.CreatorId = c.Get(GetPrincipalIdContextKey()).(int)

		database, err := s.DatabaseService.CreateDatabase(context.Background(), databaseCreate)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ECONFLICT {
				return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Database name already exists: %s", databaseCreate.Name))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create database").SetInternal(err)
		}

		if err := s.ComposeDatabaseRelationship(context.Background(), database, c.Get(getIncludeKey()).([]string)); err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal create database response").SetInternal(err)
		}
		return nil
	})

	g.GET("/database", func(c echo.Context) error {
		workspaceId := api.DEFAULT_WORKPSACE_ID
		databaseFind := &api.DatabaseFind{
			WorkspaceId: &workspaceId,
		}
		if instanceIdStr := c.QueryParam("instance"); instanceIdStr != "" {
			instanceId, err := strconv.Atoi(instanceIdStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query parameter instance is not a number: %s", instanceIdStr)).SetInternal(err)
			}
			databaseFind.InstanceId = &instanceId
		}
		list, err := s.DatabaseService.FindDatabaseList(context.Background(), databaseFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch database list").SetInternal(err)
		}

		for _, database := range list {
			if err := s.ComposeDatabaseRelationship(context.Background(), database, c.Get(getIncludeKey()).([]string)); err != nil {
				return err
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, list); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal database list response").SetInternal(err)
		}
		return nil
	})

	g.GET("/database/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		database, err := s.ComposeDatabaseById(context.Background(), id, c.Get(getIncludeKey()).([]string))
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal project ID response: %v", id)).SetInternal(err)
		}
		return nil
	})

	g.PATCH("/database/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Database ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		databasePatch := &api.DatabasePatch{
			ID:          id,
			WorkspaceId: api.DEFAULT_WORKPSACE_ID,
			UpdaterId:   c.Get(GetPrincipalIdContextKey()).(int),
		}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, databasePatch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformatted patch database request").SetInternal(err)
		}

		database, err := s.DatabaseService.PatchDatabase(context.Background(), databasePatch)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to patch database ID: %v", id)).SetInternal(err)
		}

		if err := s.ComposeDatabaseRelationship(context.Background(), database, c.Get(getIncludeKey()).([]string)); err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal database ID response: %v", id)).SetInternal(err)
		}
		return nil
	})
}

func (s *Server) ComposeDatabaseById(ctx context.Context, id int, includeList []string) (*api.Database, error) {
	databaseFind := &api.DatabaseFind{
		ID: &id,
	}
	database, err := s.DatabaseService.FindDatabase(context.Background(), databaseFind)
	if err != nil {
		if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
			return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch database ID: %v", id)).SetInternal(err)
	}

	if err := s.ComposeDatabaseRelationship(ctx, database, includeList); err != nil {
		return nil, err
	}

	return database, nil
}

func (s *Server) ComposeDatabaseRelationship(ctx context.Context, database *api.Database, includeList []string) error {
	var err error

	database.Creator, err = s.ComposePrincipalById(context.Background(), database.CreatorId, includeList)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch creator for database: %v", database.Name)).SetInternal(err)
	}

	database.Updater, err = s.ComposePrincipalById(context.Background(), database.UpdaterId, includeList)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch updater for database: %v", database.Name)).SetInternal(err)
	}

	if sort.SearchStrings(includeList, "project") >= 0 {
		database.Project, err = s.ComposeProjectlById(context.Background(), database.ProjectId, includeList)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch project for database: %v", database.Name)).SetInternal(err)
		}
	}

	if sort.SearchStrings(includeList, "instance") >= 0 {
		database.Instance, err = s.ComposeInstanceById(context.Background(), database.InstanceId, includeList)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch instance for database: %v", database.Name)).SetInternal(err)
		}
	}

	if sort.SearchStrings(includeList, "dataSource") >= 0 {
		database.DataSourceList = []*api.DataSource{}
	}

	return nil
}
