#include <cstdarg>
#include <cstdint>
#include <cstdlib>
#include <ostream>
#include <new>

/// C callback signature: void (*MessageCallback)(const char*)
using MessageCallback = void(*)(const char*);

extern "C" {

/// Start WebSocket loop inside a Tokio runtime
void tsw_controller_mod_start();

/// Stop the module
void tsw_controller_mod_stop();

/// Register callback
void tsw_controller_mod_set_receive_message_callback(MessageCallback cb);

/// Send message
void tsw_controller_mod_send_message(const char *message);

}  // extern "C"
