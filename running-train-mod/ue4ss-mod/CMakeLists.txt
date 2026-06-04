project(RunningTrainControllerMod)

set(TARGET RunningTrainControllerMod)
project(${TARGET})

set(${TARGET}_Sources
        "${CMAKE_CURRENT_SOURCE_DIR}/src/dllmain.cpp"
        )

add_library(${TARGET} SHARED ${${TARGET}_Sources})
target_include_directories(${TARGET} PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/include")
target_link_directories(${TARGET} PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/lib")
target_link_libraries(${TARGET}
    PUBLIC
        UE4SS
    PRIVATE
        tsw_controller_mod_socket_connection
)
