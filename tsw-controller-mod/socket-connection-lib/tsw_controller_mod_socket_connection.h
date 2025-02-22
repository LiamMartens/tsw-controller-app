#include <cstdarg>
#include <cstdint>
#include <cstdlib>
#include <ostream>
#include <new>

extern "C" {

void start();

void set_direct_controller_callback(void (*callback)(const char*));

void send_sync_controller_message(const char *message);

}  // extern "C"
