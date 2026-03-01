package web

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/service"
)

func RegisterRoutes(mux *http.ServeMux, db *bun.DB, userSvc service.UserServicer, cveSvc service.CveServicer, subSvc service.SubscriptionServicer) {
	h := &handler{
		db:      db,
		userSvc: userSvc,
		cveSvc:  cveSvc,
		subSvc:  subSvc,
	}

	p := "/api/v1"

	// Auth
	mux.HandleFunc(p+"/auth/login", h.enableCORS(h.apiLogin))
	mux.HandleFunc(p+"/auth/register", h.enableCORS(h.apiRegister))
	mux.HandleFunc(p+"/auth/logout", h.enableCORS(h.apiLogout))

	// User
	mux.HandleFunc(p+"/me", h.enableCORS(h.apiMe))
	mux.HandleFunc(p+"/profile", h.enableCORS(h.apiUpdateProfile))

	// Data Lists (Public)
	mux.HandleFunc(p+"/cves", h.enableCORS(h.apiCVEList))
	mux.HandleFunc(p+"/cves/", h.enableCORS(h.apiCVEDetail))
	mux.HandleFunc(p+"/vendors", h.enableCORS(h.apiVendorList))
	mux.HandleFunc(p+"/products", h.enableCORS(h.apiProductList))

	// Stats
	mux.HandleFunc(p+"/stats", h.enableCORS(h.apiStats))

	// Subscriptions (Private)
	mux.HandleFunc(p+"/subscriptions", h.enableCORS(h.apiSubscriptionManage))
	mux.HandleFunc(p+"/subscriptions/threshold", h.enableCORS(h.apiUpdateThreshold))
	mux.HandleFunc(p+"/my/dashboard", h.enableCORS(h.apiMyDashboard))

	// Admin
	mux.HandleFunc(p+"/admin/users", h.enableCORS(h.apiAdminUsers))
	mux.HandleFunc(p+"/admin/users/role", h.enableCORS(h.apiUpdateUserRole))
}

type handler struct {
	db      *bun.DB
	userSvc service.UserServicer
	cveSvc  service.CveServicer
	subSvc  service.SubscriptionServicer
}

func (h *handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *handler) enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		// When credentials are used, origin cannot be "*"
		// Use the actual request origin or a specific allowed origin
		if origin != "*" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (h *handler) getUserFromSession(r *http.Request) (*model.User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, false
	}
	user, err := h.userSvc.GetProfile(r.Context(), cookie.Value)
	if err != nil {
		return nil, false
	}
	return user, true
}

func (h *handler) apiLogin(w http.ResponseWriter, r *http.Request) {
	var req struct{ Username, Password string }
	json.NewDecoder(r.Body).Decode(&req)
	token, _, err := h.userSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		h.respondError(w, 401, "Invalid credentials")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: token, Expires: time.Now().Add(24 * time.Hour), HttpOnly: true, Path: "/", SameSite: http.SameSiteLaxMode})
	h.respondJSON(w, 200, map[string]string{"token": token})
}

func (h *handler) apiRegister(w http.ResponseWriter, r *http.Request) {
	var req struct{ Username, Email, Password string }
	json.NewDecoder(r.Body).Decode(&req)
	_, err := h.userSvc.Register(r.Context(), req.Username, req.Email, req.Password, req.Password)
	if err != nil {
		h.respondError(w, 400, err.Error())
		return
	}
	h.respondJSON(w, 201, map[string]string{"status": "ok"})
}

func (h *handler) apiLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	h.respondJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *handler) apiMe(w http.ResponseWriter, r *http.Request) {
	u, ok := h.getUserFromSession(r)
	if !ok {
		h.respondError(w, 401, "Unauthorized")
		return
	}
	h.respondJSON(w, 200, u)
}

func (h *handler) apiUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.respondError(w, 405, "Method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			h.respondError(w, 401, "Unauthorized")
			return
		}
		authHeader = "Bearer " + cookie.Value
	}

	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		h.respondError(w, 401, "Invalid authorization header")
		return
	}

	var req struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ConfirmPassword  string `json:"confirm_password"`
		NotificationFreq string `json:"notification_frequency"`
		NotificationTime string `json:"notification_time"`
		Timezone         string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "Invalid request body")
		return
	}

	// Validate notification time format (HH:MM)
	if req.NotificationTime != "" {
		if len(req.NotificationTime) != 5 || req.NotificationTime[2] != ':' {
			h.respondError(w, 400, "Invalid notification_time format. Expected HH:MM")
			return
		}
	}

	// Validate notification frequency
	if req.NotificationFreq != "" {
		validFreqs := map[string]bool{"immediate": true, "daily": true, "weekly": true}
		if !validFreqs[req.NotificationFreq] {
			h.respondError(w, 400, "Invalid notification_frequency. Must be: immediate, daily, or weekly")
			return
		}
	}

	user, err := h.userSvc.UpdateProfile(r.Context(), token, req.Email, req.Password, req.ConfirmPassword, req.NotificationFreq, req.NotificationTime, req.Timezone)
	if err != nil {
		h.respondError(w, 400, err.Error())
		return
	}

	h.respondJSON(w, 200, user)
}

func (h *handler) apiCVEList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit
	var results []struct {
		model.CVE
		SourceName string `json:"source_name"`
		Assets     string `json:"assets"`
	}
	query := h.db.NewSelect().TableExpr("cves AS c").ColumnExpr("c.*, cs.name AS source_name, STRING_AGG(DISTINCT v.name || ':' || p.name, ', ') AS assets").Join("LEFT JOIN cve_sources cs ON cs.id = c.source_id").Join("LEFT JOIN cve_products cp ON cp.cve_id = c.id").Join("LEFT JOIN products p ON p.id = cp.product_id").Join("LEFT JOIN vendors v ON v.id = p.vendor_id")
	if search != "" {
		query.Where("c.source_uid ILIKE ? OR c.title ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	count, _ := query.Group("c.id", "cs.name").Count(r.Context())
	_ = query.Order("c.published_at DESC").Limit(limit).Offset(offset).Scan(r.Context(), &results)
	h.respondJSON(w, 200, map[string]interface{}{"data": results, "total": count, "page": page, "total_pages": int(math.Ceil(float64(count) / float64(limit)))})
}

func (h *handler) apiCVEDetail(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path: /api/v1/cves/{id}
	id := r.URL.Path[len("/api/v1/cves/"):]
	if id == "" {
		h.respondError(w, 400, "CVE ID is required")
		return
	}

	// Get CVE basic info
	cve, err := h.cveSvc.Get(r.Context(), id)
	if err != nil {
		h.respondError(w, 404, "CVE not found")
		return
	}

	// Get source info
	var source model.CVESource
	err = h.db.NewSelect().Model(&source).Where("id = ?", cve.SourceID).Scan(r.Context())
	if err != nil {
		// If source not found, use empty struct
		source = model.CVESource{Name: "Unknown"}
	}

	// Get related products with vendor info
	type productWithVendor struct {
		model.Product
		VendorName string `json:"vendor_name" bun:"vendor_name"`
	}
	var products []productWithVendor
	h.db.NewSelect().
		TableExpr("products AS p").
		ColumnExpr("p.*, v.name AS vendor_name").
		Join("JOIN vendors v ON v.id = p.vendor_id").
		Join("JOIN cve_products cp ON cp.product_id = p.id").
		Where("cp.cve_id = ?", id).
		Order("v.name ASC", "p.name ASC").
		Scan(r.Context(), &products)

	// Get references if any
	var references []model.CVEReference
	h.db.NewSelect().Model(&references).Where("cve_id = ?", id).Scan(r.Context())

	// Get weaknesses if any
	var weaknesses []model.CVEWeakness
	h.db.NewSelect().Model(&weaknesses).Where("cve_id = ?", id).Scan(r.Context())

	// Construct response
	response := map[string]interface{}{
		"cve":        cve,
		"source":     source,
		"products":   products,
		"references": references,
		"weaknesses": weaknesses,
	}

	h.respondJSON(w, 200, response)
}

func (h *handler) apiVendorList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	vendorFilter := q.Get("vendor")
	productFilter := q.Get("product")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	type VendorWithCount struct {
		model.Vendor
		ProductCount   int    `json:"product_count" bun:"product_count"`
		SubscriptionID string `json:"subscription_id" bun:"subscription_id"`
	}

	// Count query
	countQuery := h.db.NewSelect().
		TableExpr("vendors AS v")

	if vendorFilter != "" {
		countQuery.Where("v.name ILIKE ?", "%"+vendorFilter+"%")
	}
	if productFilter != "" {
		countQuery.Where("EXISTS (SELECT 1 FROM products WHERE vendor_id = v.id AND name ILIKE ?)", "%"+productFilter+"%")
	}

	count, _ := countQuery.Count(r.Context())

	// Data query with subqueries
	var results []VendorWithCount
	dataQuery := h.db.NewSelect().
		TableExpr("vendors AS v").
		ColumnExpr("v.*").
		ColumnExpr("(SELECT COUNT(*) FROM products WHERE vendor_id = v.id) as product_count").
		ColumnExpr("(SELECT st.subscription_id FROM subscription_targets st WHERE st.target_type_id = 2 AND st.target_id = v.id LIMIT 1) as subscription_id")

	if vendorFilter != "" {
		dataQuery.Where("v.name ILIKE ?", "%"+vendorFilter+"%")
	}
	if productFilter != "" {
		dataQuery.Where("EXISTS (SELECT 1 FROM products WHERE vendor_id = v.id AND name ILIKE ?)", "%"+productFilter+"%")
	}

	_ = dataQuery.Order("v.name ASC").Limit(limit).Offset(offset).Scan(r.Context(), &results)

	h.respondJSON(w, 200, map[string]interface{}{
		"data":        results,
		"total":       count,
		"page":        page,
		"total_pages": int(math.Ceil(float64(count) / float64(limit))),
	})
}

func (h *handler) apiProductList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	productFilter := q.Get("product")
	vendorFilter := q.Get("vendor")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	type ProductWithVendor struct {
		model.Product
		VendorName     string `json:"vendor_name" bun:"vendor_name"`
		SubscriptionID string `json:"subscription_id" bun:"subscription_id"`
	}

	// Count query
	countQuery := h.db.NewSelect().
		TableExpr("products AS p").
		Join("LEFT JOIN vendors v ON v.id = p.vendor_id")

	if productFilter != "" {
		countQuery.Where("p.name ILIKE ?", "%"+productFilter+"%")
	}
	if vendorFilter != "" {
		countQuery.Where("v.name ILIKE ?", "%"+vendorFilter+"%")
	}

	count, _ := countQuery.Count(r.Context())

	// Data query with subquery for subscription
	var results []ProductWithVendor
	dataQuery := h.db.NewSelect().
		TableExpr("products AS p").
		ColumnExpr("p.*").
		ColumnExpr("v.name as vendor_name").
		ColumnExpr("(SELECT st.subscription_id FROM subscription_targets st WHERE st.target_type_id = 3 AND st.target_id = p.id LIMIT 1) as subscription_id").
		Join("LEFT JOIN vendors v ON v.id = p.vendor_id")

	if productFilter != "" {
		dataQuery.Where("p.name ILIKE ?", "%"+productFilter+"%")
	}
	if vendorFilter != "" {
		dataQuery.Where("v.name ILIKE ?", "%"+vendorFilter+"%")
	}

	_ = dataQuery.Order("p.name ASC").Limit(limit).Offset(offset).Scan(r.Context(), &results)

	h.respondJSON(w, 200, map[string]interface{}{
		"data":        results,
		"total":       count,
		"page":        page,
		"total_pages": int(math.Ceil(float64(count) / float64(limit))),
	})
}

func (h *handler) apiStats(w http.ResponseWriter, r *http.Request) {
	tc, _ := h.db.NewSelect().Model((*model.CVE)(nil)).Count(r.Context())
	tv, _ := h.db.NewSelect().Model((*model.Vendor)(nil)).Count(r.Context())
	tp, _ := h.db.NewSelect().Model((*model.Product)(nil)).Count(r.Context())

	// Generate chart data based on range parameter
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "1m"
	}

	// Get max date from daily_cve_counts table
	type MaxDayResult struct {
		MaxDay time.Time `bun:"max_day"`
	}
	var maxDayResult MaxDayResult
	err := h.db.NewSelect().
		Table("daily_cve_counts").
		ColumnExpr("MAX(day) as max_day").
		Scan(r.Context(), &maxDayResult)

	end := maxDayResult.MaxDay
	if err != nil || end.IsZero() {
		end = time.Now()
	}

	// Calculate start date based on time range
	var start time.Time
	switch timeRange {
	case "1w":
		start = end.AddDate(0, 0, -7)
	case "1m":
		start = end.AddDate(0, -1, 0)
	case "1y":
		start = end.AddDate(-1, 0, 0)
	case "5y":
		start = end.AddDate(-5, 0, 0)
	default:
		start = end.AddDate(0, -1, 0) // default 1 month
	}

	type ChartRow struct {
		Period time.Time `bun:"period"`
		Count  int       `bun:"count"`
	}
	var rows []ChartRow
	dateMap := make(map[string]int)
	var chartData []map[string]interface{}

	// Different aggregation strategy based on time range
	if timeRange == "5y" {
		// Monthly aggregation for 5 years
		err := h.db.NewSelect().
			Table("daily_cve_counts").
			ColumnExpr("DATE_TRUNC('month', day) as period, SUM(count) as count").
			Where("day >= ? AND day <= ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
			Group("period").
			Order("period ASC").
			Scan(r.Context(), &rows)

		if err == nil {
			for _, row := range rows {
				dateMap[row.Period.UTC().Format("2006-01-01")] = row.Count
			}

			// Fill all months
			for d := start; !d.After(end); {
				firstOfMonth := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
				key := firstOfMonth.Format("2006-01-01")
				chartData = append(chartData, map[string]interface{}{
					"period": firstOfMonth.Format(time.RFC3339),
					"count":  dateMap[key],
				})
				d = firstOfMonth.AddDate(0, 1, 0)
			}
		}
	} else if timeRange == "1y" {
		// Weekly aggregation for 1 year
		alignToMonday := func(t time.Time) time.Time {
			wd := int(t.Weekday())
			if wd == 0 {
				wd = 7
			}
			return t.AddDate(0, 0, -(wd - 1))
		}
		queryStart := alignToMonday(start)

		err := h.db.NewSelect().
			Table("daily_cve_counts").
			ColumnExpr("DATE_TRUNC('week', day) as period, SUM(count) as count").
			Where("day >= ? AND day <= ?", queryStart.Format("2006-01-02"), end.Format("2006-01-02")).
			Group("period").
			Order("period ASC").
			Scan(r.Context(), &rows)

		if err == nil {
			for _, row := range rows {
				dateMap[row.Period.UTC().Format("2006-01-02")] = row.Count
			}

			// Fill all weeks
			for d := queryStart; !d.After(end); d = d.AddDate(0, 0, 7) {
				key := d.Format("2006-01-02")
				chartData = append(chartData, map[string]interface{}{
					"period": d.Format(time.RFC3339),
					"count":  dateMap[key],
				})
			}
		}
	} else {
		// Daily aggregation for 1w and 1m
		err := h.db.NewSelect().
			Table("daily_cve_counts").
			ColumnExpr("day as period, count").
			Where("day >= ? AND day <= ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
			Order("period ASC").
			Scan(r.Context(), &rows)

		if err == nil {
			for _, row := range rows {
				dateMap[row.Period.UTC().Format("2006-01-02")] = row.Count
			}

			// Fill all days
			for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
				key := d.Format("2006-01-02")
				chartData = append(chartData, map[string]interface{}{
					"period": d.Format(time.RFC3339),
					"count":  dateMap[key],
				})
			}
		}
	}

	// Initialize empty array if no data
	if chartData == nil {
		chartData = make([]map[string]interface{}, 0)
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"total_cves":      tc,
		"total_vendors":   tv,
		"total_products":  tp,
		"chart_data":      chartData,
		"latest_data_date": end.Format("2006-01-02"),
	})
}

func (h *handler) apiSubscriptionManage(w http.ResponseWriter, r *http.Request) {
	u, ok := h.getUserFromSession(r)
	if !ok {
		h.respondError(w, 401, "Unauthorized")
		return
	}

	if r.Method == http.MethodGet {
		// Get all user subscriptions with details
		subs, err := h.subSvc.ListSubscriptions(r.Context(), u.ID.String())
		if err != nil {
			h.respondError(w, 500, "Failed to fetch subscriptions")
			return
		}

		// Enrich with target details
		type enrichedSub struct {
			ID                string `json:"id"`
			TargetType        string `json:"target_type"`
			TargetID          string `json:"target_id"`
			TargetName        string `json:"target_name"`
			SeverityThreshold string `json:"severity_threshold"`
			CreatedAt         string `json:"created_at"`
		}

		result := make([]enrichedSub, 0)
		for _, sub := range subs {
			for _, target := range sub.Targets {
				var targetType, targetName string
				switch target.TargetTypeID {
				case 1:
					targetType = "cve_source"
					var name string
					h.db.NewSelect().Table("cve_sources").Column("name").Where("id = ?", target.TargetID).Scan(r.Context(), &name)
					targetName = name
				case 2:
					targetType = "vendor"
					var name string
					h.db.NewSelect().Table("vendors").Column("name").Where("id = ?", target.TargetID).Scan(r.Context(), &name)
					targetName = name
				case 3:
					targetType = "product"
					var name string
					h.db.NewSelect().Table("products").Column("name").Where("id = ?", target.TargetID).Scan(r.Context(), &name)
					targetName = name
				}

				result = append(result, enrichedSub{
					ID:                sub.ID.String(),
					TargetType:        targetType,
					TargetID:          target.TargetID.String(),
					TargetName:        targetName,
					SeverityThreshold: h.subSvc.SeverityToString(sub.SeverityThreshold),
					CreatedAt:         sub.CreatedAt.Format(time.RFC3339),
				})
			}
		}

		h.respondJSON(w, 200, result)

	} else if r.Method == http.MethodPost {
		var req struct {
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
			Severity   string `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW, INFO
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondError(w, 400, "Invalid request body")
			return
		}

		// Default to MEDIUM if not specified
		if req.Severity == "" {
			req.Severity = "MEDIUM"
		}

		sub, err := h.subSvc.CreateSubscription(r.Context(), u.ID.String(), []service.SubscriptionTarget{{TargetType: req.TargetType, TargetID: req.TargetID}}, req.Severity)
		if err != nil {
			h.respondError(w, 500, "Failed to create subscription")
			return
		}

		h.respondJSON(w, 201, map[string]interface{}{
			"status": "ok",
			"id":     sub.ID.String(),
		})

	} else if r.Method == http.MethodDelete {
		subID := r.URL.Query().Get("id")
		if subID == "" {
			h.respondError(w, 400, "Missing subscription id")
			return
		}

		if err := h.subSvc.DeleteSubscription(r.Context(), u.ID.String(), subID); err != nil {
			h.respondError(w, 500, "Failed to delete subscription")
			return
		}

		h.respondJSON(w, 200, map[string]string{"status": "ok"})

	} else {
		h.respondError(w, 405, "Method not allowed")
	}
}

func (h *handler) apiUpdateThreshold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		h.respondError(w, 405, "Method not allowed")
		return
	}

	u, ok := h.getUserFromSession(r)
	if !ok {
		h.respondError(w, 401, "Unauthorized")
		return
	}

	var req struct {
		ID        string `json:"id"`
		Threshold string `json:"threshold"` // CRITICAL, HIGH, MEDIUM, LOW, INFO
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "Invalid request body")
		return
	}

	if req.ID == "" || req.Threshold == "" {
		h.respondError(w, 400, "Missing id or threshold")
		return
	}

	if err := h.subSvc.UpdateThreshold(r.Context(), u.ID.String(), req.ID, req.Threshold); err != nil {
		h.respondError(w, 500, "Failed to update threshold")
		return
	}

	h.respondJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *handler) apiMyDashboard(w http.ResponseWriter, r *http.Request) {
	u, ok := h.getUserFromSession(r)
	if !ok {
		h.respondError(w, 401, "Unauthorized")
		return
	}
	type sub struct {
		ID                string `json:"id"`
		TargetType        string `json:"target_type"`
		TargetName        string `json:"target_name"`
		SeverityThreshold int16  `json:"-" bun:"severity_threshold"`
		SeverityString    string `json:"severity_threshold" bun:"-"`
	}
	v, p := make([]sub, 0), make([]sub, 0)
	h.db.NewSelect().TableExpr("subscriptions AS s").ColumnExpr("s.id, 'vendor' as target_type, v.name as target_name, s.severity_threshold").Join("JOIN subscription_targets st ON st.subscription_id = s.id").Join("JOIN vendors v ON v.id = st.target_id").Where("s.user_id = ? AND st.target_type_id = 2", u.ID).Order("s.created_at ASC", "v.name ASC").Scan(r.Context(), &v)
	h.db.NewSelect().TableExpr("subscriptions AS s").ColumnExpr("s.id, 'product' as target_type, p.name as target_name, s.severity_threshold").Join("JOIN subscription_targets st ON st.subscription_id = s.id").Join("JOIN products p ON p.id = st.target_id").Where("s.user_id = ? AND st.target_type_id = 3", u.ID).Order("s.created_at ASC", "p.name ASC").Scan(r.Context(), &p)

	// Convert severity levels to strings
	for i := range v {
		v[i].SeverityString = h.subSvc.SeverityToString(v[i].SeverityThreshold)
	}
	for i := range p {
		p[i].SeverityString = h.subSvc.SeverityToString(p[i].SeverityThreshold)
	}

	h.respondJSON(w, 200, map[string]interface{}{"vendor_subs": v, "product_subs": p})
}

func (h *handler) apiAdminUsers(w http.ResponseWriter, r *http.Request) {
	u, ok := h.getUserFromSession(r)
	if !ok || u.Role != "admin" {
		h.respondError(w, 403, "Forbidden")
		return
	}
	var users []model.User
	h.db.NewSelect().Model(&users).Order("username ASC").Scan(r.Context())
	h.respondJSON(w, 200, users)
}

func (h *handler) apiUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	u, ok := h.getUserFromSession(r)
	if !ok || u.Role != "admin" {
		h.respondError(w, 403, "Forbidden")
		return
	}
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	h.db.NewUpdate().Table("users").Set("role = ?", req.Role).Where("id = ?", req.UserID).Exec(r.Context())
	h.respondJSON(w, 200, map[string]string{"status": "ok"})
}
