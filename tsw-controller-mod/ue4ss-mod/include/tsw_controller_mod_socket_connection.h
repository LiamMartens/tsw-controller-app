#include <cstdarg>
#include <cstdint>
#include <cstdlib>
#include <ostream>
#include <new>

extern "C" {

void tsw_controller_mod_start();

void tsw_controller_mod_set_direct_controller_callback(void (*callback)(const char*));

void tsw_controller_mod_send_sync_controller_message(const char *message);

}  // extern "C"
