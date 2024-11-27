package lib

// LinkWithGithub redirects the user to GitHub for app installation.
func LinkWithGithub(c *gin.Context, user models.User) {
	appName := "byteport-gh"
	redirectURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", appName, user.UUID)
	c.Redirect(http.StatusFound, redirectURL)
}