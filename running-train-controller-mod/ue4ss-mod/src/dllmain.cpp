#include <string>
#include <format>
#include <mutex>
#include <queue>
#include <cmath>
#include <tuple>
#include <shared_mutex>
#include <unordered_map>
#include <regex>

#include <Unreal/Core/HAL/Platform.hpp>
#include <Unreal/FFrame.hpp>
#include <Unreal/FURL.hpp>
#include <Unreal/FWorldContext.hpp>
#include <Unreal/FOutputDevice.hpp>
#include <Unreal/FProperty.hpp>
#include <Unreal/Hooks.hpp>
#include <Unreal/PackageName.hpp>
#include <Unreal/Property/FArrayProperty.hpp>
#include <Unreal/Property/FBoolProperty.hpp>
#include <Unreal/Property/FClassProperty.hpp>
#include <Unreal/Property/FEnumProperty.hpp>
#include <Unreal/Property/FMapProperty.hpp>
#include <Unreal/Property/FNameProperty.hpp>
#include <Unreal/Property/FObjectProperty.hpp>
#include <Unreal/Property/FStrProperty.hpp>
#include <Unreal/Property/FStructProperty.hpp>
#include <Unreal/Property/FTextProperty.hpp>
#include <Unreal/Property/FWeakObjectProperty.hpp>
#include <Unreal/Property/NumericPropertyTypes.hpp>
#include <Unreal/TypeChecker.hpp>
#include <Unreal/UAssetRegistry.hpp>
#include <Unreal/UAssetRegistryHelpers.hpp>
#include <Unreal/UClass.hpp>
#include <Unreal/UFunction.hpp>
#include <Unreal/UGameViewportClient.hpp>
#include <Unreal/UKismetSystemLibrary.hpp>
#include <Unreal/UObjectGlobals.hpp>
#include <Unreal/UPackage.hpp>
#include <Unreal/UScriptStruct.hpp>
#include <Unreal/GameplayStatics.hpp>
#include <DynamicOutput/Output.hpp>
#include <UE4SSProgram.hpp>

#include "tsw_controller_mod_socket_connection.h"
struct EmptyVoidFunctionParams{};
struct PlayerController_IsPlayerControllerParams
{
    bool IsPlayerController;
};
struct DirectControlControlTargetState
{
    float TargetValue;
    float MaxChangeRate;
    std::vector<RC::StringType> Flags;
};

class RunningTrainControllerMod : public RC::CppUserModBase
{
  private:
    static inline std::wregex RX_SIDE_PLACEHOLDER = std::wregex(STR(R"(\{SIDE(:[^:]+)?(:[^:]+)?\})"));

    static inline std::shared_mutex CURRENT_DRIVABLE_ACTOR_CLASS_NAME_MUTEX;
    static inline float TIME_SINCE_CURRENT_DRIVABLE_ACTOR_REPORTED = 0;
    static inline RC::StringType CURRENT_DRIVABLE_ACTOR_CLASS_NAME = STR("");

    /* map of control names and their target value and flags; for running train only the controls: "Reverser", "Throttle" and "Brake" will be controllable */
    static inline std::shared_mutex DIRECT_CONTROL_TARGET_STATE_MUTEX;
    static inline std::unordered_map<RC::StringType, DirectControlControlTargetState> DIRECT_CONTROL_TARGET_STATE;


    static bool is_within_margin_of_error(float current, float target)
    {
        return abs(target - current) < 0.05f;
    }

    static bool is_player_controller(Unreal::UObject* controller)
    {
        if (!controller) return false;
        PlayerController_IsPlayerControllerParams is_player_controller_result;
        Unreal::UFunction* is_player_function = controller->GetFunctionByNameInChain(STR("IsPlayerController"));
        if (is_player_function)
        {
            controller->ProcessEvent(is_player_function, &is_player_controller_result);
            return is_player_controller_result.IsPlayerController;
        }
        return false;
    }

    static std::vector<RC::StringType> wstring_split(RC::StringType s, RC::StringType delimiter)
    {
        size_t pos_start = 0, pos_end, delim_len = delimiter.length();
        RC::StringType token;
        std::vector<RC::StringType> res;

        while ((pos_end = s.find(delimiter, pos_start)) != RC::StringType::npos)
        {
            token = s.substr(pos_start, pos_end - pos_start);
            pos_start = pos_end + delim_len;
            res.push_back(token);
        }

        res.push_back(s.substr(pos_start));
        return res;
    }

    static void on_tick(Unreal::AActor* controller, float delta_secs)
    {
        if (!RunningTrainControllerMod::is_player_controller(controller)) return;
    }

    static void on_direct_control_message_received(const char* raw_message)
    {
        /* quick nullptr check */
        if (!raw_message) return;

        auto message = RC::ensure_str(std::string{raw_message});
        auto parts = RunningTrainControllerMod::wstring_split(message, STR(","));

        /* update DC target state */
        std::unique_lock<std::shared_mutex> lock(RunningTrainControllerMod::DIRECT_CONTROL_TARGET_STATE_MUTEX);
        /* format: direct_control,controls={control_name},value={target_value},max_change_rate={max_rate},flags={flag|flag} */
        if (parts[0] != STR("direct_control")) return;
        std::map<RC::StringType, RC::StringType> properties;
        for (size_t i = 1; i < parts.size(); ++i)
        {
            const RC::StringType& kv = parts[i];
            size_t eqPos = kv.find(STR("="));
            if (eqPos != RC::StringType::npos) {
                auto key = kv.substr(0, eqPos);
                auto val = kv.substr(eqPos + 1);
                properties[key] = val;
            }
        }

        /* skip for missing properties */
        if (properties.find(STR("value")) == properties.end()) return;
        if (properties.find(STR("flags")) == properties.end()) return;

        Output::send<LogLevel::Normal>(STR("[RunningTrainControllerMod] Processing Direct Control Message: {}\n"), message);
        float target_value = std::stof(properties[STR("value")]);
        float max_change_rate = (properties.find(STR("max_change_rate")) == properties.end()) ? 1000.0 : std::stof(properties[STR("max_change_rate")]);
        std::vector<RC::StringType> flags = RunningTrainControllerMod::wstring_split(properties[STR("flags")], STR("|"));
        DirectControlControlTargetState control_target_state{target_value,max_change_rate,flags};
        RunningTrainControllerMod::DIRECT_CONTROL_TARGET_STATE[properties[STR("controls")]] = control_target_state;
    }

  public:
    RunningTrainControllerMod() : CppUserModBase()
    {
        ModName = STR("RunningTrainControllerMod");
        ModVersion = STR("1.10.1");
        ModDescription = STR("TSW Controller Utility Helper");
        ModAuthors = STR("Liah Martens");

        Output::send<LogLevel::Normal>(STR("[RunningTrainControllerMod] Starting...\n"));
    }

    auto on_unreal_init() -> void override
    {
        Output::send<LogLevel::Normal>(STR("[RunningTrainControllerMod] Unreal Initialized\n"));

        Unreal::Hook::RegisterAActorTickPreCallback(RunningTrainControllerMod::on_tick);
        tsw_controller_mod_set_receive_message_callback(RunningTrainControllerMod::on_direct_control_message_received);
    }

    ~RunningTrainControllerMod() override = default;
};

#define RUNNING_TRAIN_CONTROLLER_MOD_API __declspec(dllexport)
extern "C"
{
    RUNNING_TRAIN_CONTROLLER_MOD_API RC::CppUserModBase* start_mod()
    {
        tsw_controller_mod_start();
        return new RunningTrainControllerMod();
    }

    RUNNING_TRAIN_CONTROLLER_MOD_API void uninstall_mod(RC::CppUserModBase* mod)
    {
        tsw_controller_mod_stop();
        delete mod;
    }
}
