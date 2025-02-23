local projectName = "TSWControllerMod"

target(projectName)
  add_rules("ue4ss.mod")
  add_includedirs("include")
  add_linkdirs("lib")
  add_links("tsw_controller_mod_socket_connection")
  add_files("src/dllmain.cpp")