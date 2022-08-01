#include "gliblog.h"

static GLogLevelFlags all_levels =
  G_LOG_FLAG_RECURSION |
  G_LOG_FLAG_FATAL |
  G_LOG_LEVEL_ERROR |
  G_LOG_LEVEL_CRITICAL |
  G_LOG_LEVEL_WARNING;

void
log_handler(const gchar *log_domain, GLogLevelFlags log_level,
  const gchar *message, gpointer user_data) {

  logGLib((char *)log_domain, log_level, (char *)message);
}

void
glib_log_configure() {
  g_log_set_handler (NULL, all_levels, log_handler, NULL);
  g_log_set_handler ("VIPS",  all_levels, log_handler, NULL);
}
