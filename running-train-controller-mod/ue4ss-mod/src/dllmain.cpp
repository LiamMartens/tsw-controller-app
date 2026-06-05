#include <string>
#include <format>
#include <mutex>
#include <queue>
#include <cmath>
#include <tuple>
#include <shared_mutex>
#include <unordered_map>
#include <regex>
#include <sstream>

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

const TCHAR* CONTROL_NAME_THROTTLE = L"Throttle";
const TCHAR* CONTROL_NAME_REVERSER = L"Reverser";
const TCHAR* CONTROL_NAME_BRAKE = L"Brake";

struct EmptyVoidFunctionParams{};
struct Actor_IsPlayerControlledParams
{
    bool IsPlayerControlled;
};

struct DirectControlControlTargetState
{
    float TargetValue;
    float MaxChangeRate;
    std::vector<RC::StringType> Flags;
};

class RunningTrainState
{
public:
    RC::StringType ClassName = STR("");
    int Throttle = 0;
    int Reverser = 0;
    int Brake = 0;

    void SetClassName(RC::StringType name) { this->ClassName = name; }
    void SetThrottle(int throttle) { this->Throttle = throttle; }
    void SetReverser(int reverser) { this->Reverser = reverser; }
    void SetBrake(int brake) { this->Brake = brake; }
};

class RunningTrainControllerMod : public RC::CppUserModBase
{
  private:
    static inline std::wregex RX_SIDE_PLACEHOLDER = std::wregex(STR(R"(\{SIDE(:[^:]+)?(:[^:]+)?\})"));

    static inline float TIME_SINCE_TRAIN_STATE_REPORTED = 0;
    static inline std::shared_mutex TRAIN_STATE_MUTEX;
    static inline std::shared_ptr<RunningTrainState> TRAIN_STATE = nullptr;

    /* map of control names and their target value and flags; for running train only the controls: "Reverser", "Throttle" and "Brake" will be controllable */
    static inline std::shared_mutex DIRECT_CONTROL_TARGET_STATE_MUTEX;
    static inline std::unordered_map<RC::StringType, DirectControlControlTargetState> DIRECT_CONTROL_TARGET_STATE;


    static bool is_within_margin_of_error(float current, float target)
    {
        return abs(target - current) < 0.05f;
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

    static bool is_player_controlled(Unreal::UObject* actor)
    {
        if (!actor) return false;
        Actor_IsPlayerControlledParams is_player_controlled_result;
        Unreal::UFunction* is_player_function = actor->GetFunctionByNameInChain(STR("IsPlayerControlled"));
        if (is_player_function)
        {
            actor->ProcessEvent(is_player_function, &is_player_controlled_result);
            return is_player_controlled_result.IsPlayerControlled;
        }
        return false;
    }

    static Unreal::UObject* get_base_train_from_pawn(Unreal::UObject* pawn)
    {
        Unreal::FObjectProperty* base_train_prop =
                static_cast<Unreal::FObjectProperty*>(pawn->GetPropertyByNameInChain(STR("BaseTrain")));
        if (!base_train_prop) return nullptr;
        void* base_train_addr = base_train_prop->ContainerPtrToValuePtr<void>(pawn);
        Unreal::UObject* base_train = base_train_prop->GetObjectPropertyValue(base_train_addr);
        if (!base_train) return nullptr;
        return base_train;
    }

    static int get_int_value_from_property(Unreal::UObject* obj, const TCHAR* prop_name)
    {
        Unreal::FProperty* prop = static_cast<Unreal::FProperty*>(obj->GetPropertyByNameInChain(prop_name));
        if (!prop) return 0;
        int* value =  prop->ContainerPtrToValuePtr<int>(obj);
        return *value;
    }

    static void set_int_value_from_property(Unreal::UObject* obj, const TCHAR* prop_name, int incoming)
    {
        Unreal::FProperty* prop = static_cast<Unreal::FProperty*>(obj->GetPropertyByNameInChain(prop_name));
        if (prop)
        {
            int* value =  prop->ContainerPtrToValuePtr<int>(obj);
            *value = incoming;
        }
    }

    static void send_sync_control_value(const TCHAR* control_name, int value)
    {
            std::wstringstream message;
            message << STR("sync_control_value,name=") << control_name << STR(",property=") << control_name
                    << STR(",value=") << std::to_wstring(value) << STR(",normalized_value=") << std::to_wstring(value);
            auto message_wstr = message.str();
            auto message_str = std::string(message_wstr.begin(), message_wstr.end());
            Output::send<LogLevel::Normal>(STR("[TSWControllerMod] sending updated control value: {}\n"), message_wstr);
            tsw_controller_mod_send_message((char*)message_str.c_str());
    }

    static void on_tick(Unreal::AActor* actor, float delta_secs)
    {
        if (!RunningTrainControllerMod::is_player_controlled(actor)) return;

        auto base_train = RunningTrainControllerMod::get_base_train_from_pawn(actor);
        if (!base_train) {
            std::unique_lock<std::shared_mutex> train_state_lock(RunningTrainControllerMod::DIRECT_CONTROL_TARGET_STATE_MUTEX);
            if (RunningTrainControllerMod::TRAIN_STATE)
            {
                /* reset current actor if a train state was set */
                RC::StringType message = STR("current_drivable_actor,name=");
                auto message_str = std::string(message.begin(), message.end());
                tsw_controller_mod_send_message((char*)message_str.c_str());
            }
            RunningTrainControllerMod::TRAIN_STATE = nullptr;
            return;
        };

        auto base_train_name = base_train->GetClassPrivate()->GetName();
        std::unique_lock<std::shared_mutex> train_state_lock(RunningTrainControllerMod::TRAIN_STATE_MUTEX);
        if (!RunningTrainControllerMod::TRAIN_STATE) { RunningTrainControllerMod::TRAIN_STATE = std::make_shared<RunningTrainState>(); }

        /* update train state*/
        RunningTrainControllerMod::TRAIN_STATE->SetClassName(base_train_name);
        RunningTrainControllerMod::TRAIN_STATE->SetReverser(RunningTrainControllerMod::get_int_value_from_property(actor, STR("CR_Reverser")));
        RunningTrainControllerMod::TRAIN_STATE->SetThrottle(RunningTrainControllerMod::get_int_value_from_property(actor, STR("CP_Throttle")));
        RunningTrainControllerMod::TRAIN_STATE->SetBrake(RunningTrainControllerMod::get_int_value_from_property(actor, STR("CB Brake")));

        RunningTrainControllerMod::TIME_SINCE_TRAIN_STATE_REPORTED += delta_secs;
        if (
            RunningTrainControllerMod::TRAIN_STATE->ClassName != base_train_name ||
            RunningTrainControllerMod::TIME_SINCE_TRAIN_STATE_REPORTED > 1.0f
        ) {
            RunningTrainControllerMod::TIME_SINCE_TRAIN_STATE_REPORTED = 0.0f;
            auto message = STR("current_drivable_actor,name=") + base_train_name;
            auto message_str = std::string(message.begin(), message.end());
            Output::send<LogLevel::Normal>(STR("[RunningTrainControllerMod] sending current drivable actor information {}\n"), message);
            tsw_controller_mod_send_message((char*)message_str.c_str());
            RunningTrainControllerMod::send_sync_control_value(CONTROL_NAME_THROTTLE, RunningTrainControllerMod::TRAIN_STATE->Throttle);
            RunningTrainControllerMod::send_sync_control_value(CONTROL_NAME_REVERSER, RunningTrainControllerMod::TRAIN_STATE->Reverser);
            RunningTrainControllerMod::send_sync_control_value(CONTROL_NAME_BRAKE, RunningTrainControllerMod::TRAIN_STATE->Brake);
        }

        std::unique_lock<std::shared_mutex> direct_control_target_state_lock(RunningTrainControllerMod::DIRECT_CONTROL_TARGET_STATE_MUTEX);
        for (const auto& control_pair : RunningTrainControllerMod::DIRECT_CONTROL_TARGET_STATE)
        {
            auto target_value = control_pair.second.TargetValue;
            auto flags = control_pair.second.Flags;
            bool should_hold = std::find(flags.begin(), flags.end(), STR("hold")) != flags.end();
            bool should_be_relative = std::find(flags.begin(), flags.end(), STR("relative")) != flags.end();

            if (control_pair.first == CONTROL_NAME_THROTTLE)
            {
                if (should_be_relative) { target_value = RunningTrainControllerMod::TRAIN_STATE->Throttle + target_value; }
                RunningTrainControllerMod::set_int_value_from_property(actor, STR("CP_Throttle"), target_value);
                RunningTrainControllerMod::set_int_value_from_property(base_train, STR("P_Throttle"), target_value);
                RunningTrainControllerMod::set_int_value_from_property(base_train, STR("Reg_Throttle"), target_value);
            }

            if (control_pair.first == CONTROL_NAME_REVERSER)
            {
                if (should_be_relative) { target_value = RunningTrainControllerMod::TRAIN_STATE->Reverser + target_value; }
                RunningTrainControllerMod::set_int_value_from_property(actor, STR("CR_Reverser"), target_value);
                RunningTrainControllerMod::set_int_value_from_property(base_train, STR("P_Reverser"), target_value);
            }

            if (control_pair.first == CONTROL_NAME_BRAKE)
            {
                if (should_be_relative) { target_value = RunningTrainControllerMod::TRAIN_STATE->Brake + target_value; }
                RunningTrainControllerMod::set_int_value_from_property(actor, STR("CB Brake"), target_value);
                RunningTrainControllerMod::set_int_value_from_property(base_train, STR("P_Brake"), target_value);
                RunningTrainControllerMod::set_int_value_from_property(base_train, STR("Reg_Brake"), target_value);
            }

            if (!should_hold)
            {
                RunningTrainControllerMod::DIRECT_CONTROL_TARGET_STATE.erase(control_pair.first);
            }
        }
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
        ModVersion = STR("1.17.0");
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
