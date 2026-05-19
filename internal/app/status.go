package app

import "context"

func (s *Service) GetStatus(ctx context.Context, telegramUserID int64) StatusReport {
	backendStatus := ProbeBackendLatency(ctx, s.backend)
	report := StatusReport{
		ServerURL:        s.serverURL,
		DataFile:         s.dataFile,
		BackendLatency:   backendStatus.Latency,
		BackendAvailable: backendStatus.Err == nil,
		LinkedUsers:      s.store.CountUserAccessTokens(),
	}
	if backendStatus.Err != nil {
		report.BackendError = sanitizeBackendError(backendStatus.Err)
	}

	if s.instanceProfile != nil {
		report.InstanceURL = s.instanceProfile.InstanceURL
	}

	if len(s.allowedUsernames) > 0 {
		report.AllowedUsernames = len(s.allowedUsernames)
	}

	accessToken, ok := s.store.GetUserAccessToken(telegramUserID)
	if !ok {
		return report
	}
	report.AccountLinked = true

	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return report
	}
	report.AccountTokenValid = true
	report.AccountDisplayName = displayNameOf(user)
	return report
}
