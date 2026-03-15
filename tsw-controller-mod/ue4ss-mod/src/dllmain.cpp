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
struct VirtualHIDComponent_GetCurrentlyChangingControllerParams
{
    Unreal::UObject* Controller;
};
struct VirtualHIDComponent_InputValueChangedParams
{
    float OldValue;
    float NewValue;
};
struct PlayerController_IsPlayerControllerParams
{
    bool IsPlayerController;
};
struct PlayerController_GetDriverPawnParams
{
    Unreal::UObject* DriverPawn;
};
struct DriverPawn_GetAttachedSeatComponentParams
{
    Unreal::UObject* SeatComponent;
};
struct DriverController_GetDrivableActorParams
{
    Unreal::UObject* DrivableActor;
};
struct RailVehicle_FindVirtualHIDComponentParams
{
    Unreal::FName Name;
    Unreal::UObject* VirtualHIDComponent;
};
struct VirtualHIDComponent_SetCurrentInputValueParams
{
    float Value;
};
struct VirtualHIDComponent_SetNormalisedInputValueParams
{
    float Value;
};
struct VirtualHIDComponent_SetPushedStateParams
{
    bool IsPushed;
    bool SkipTransition;
};
struct VirtualHIDComponent_IsChangingParams
{
    bool IsChanging;
};
struct VirtualHIDComponent_GetCurrentInputValueParams
{
    float InputValue;
};
struct VirtualVHIDComponent_GetNormalisedInputValueParams
{
    float InputValue;
};
struct VirtualVHIDComponent_EndChangingParams
{
    Unreal::UObject* Controller;
    bool Result;
};
struct Controller_NotifyBeginInteractionParams
{
    Unreal::UObject* Component;
};
struct PlayerController_BeginChangingVHIDComponentParams
{
    Unreal::UObject* Component;
};
struct PlayerController_EndUsingVHIDComponentParams
{
    Unreal::UObject* Component;
};

struct DirectControlControlTargetState
{
    float TargetValue;
    float MaxChangeRate;
    std::vector<RC::StringType> Flags;
};

class TSWControllerMod : public RC::CppUserModBase
{
  private:
    static inline std::shared_mutex CURRENT_DRIVABLE_ACTOR_CLASS_NAME_MUTEX;
    static inline float TIME_SINCE_CURRENT_DRIVABLE_ACTOR_REPORTED = 0;
    static inline RC::StringType CURRENT_DRIVABLE_ACTOR_CLASS_NAME = STR("");

    /* map of control names and their target value and flags */
    static inline std::shared_mutex DIRECT_CONTROL_TARGET_STATE_MUTEX;
    static inline std::unordered_map<RC::StringType, DirectControlControlTargetState> DIRECT_CONTROL_TARGET_STATE;

    static inline std::shared_mutex VHID_COMPONENTS_CHANGING_MUTEX;
    static inline std::unordered_map<RC::StringType, Unreal::TWeakObjectPtr<Unreal::UObject>> VHID_COMPONENTS_CHANGING;

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

    static Unreal::UObject* get_driver_pawn_from_controller(Unreal::UObject* controller)
    {
        if (!controller) return nullptr;

        Unreal::UFunction* get_driver_pawn_func = controller->GetFunctionByNameInChain(STR("GetDriverPawn"));
        if (!get_driver_pawn_func) return nullptr;

        PlayerController_GetDriverPawnParams get_driver_pawn_result;
        controller->ProcessEvent(get_driver_pawn_func, &get_driver_pawn_result);

        return get_driver_pawn_result.DriverPawn;
    }

    static RC::StringType format_direct_control_name(Unreal::UObject* pawn, RC::StringType control_name)
    {
        uint8_t train_side = 0;

        /* get seat side to determine train side */
        DriverPawn_GetAttachedSeatComponentParams get_attached_seat_component_result;
        pawn->ProcessEvent(pawn->GetFunctionByNameInChain(STR("GetAttachedSeatComponent")), &get_attached_seat_component_result);
        if (get_attached_seat_component_result.SeatComponent)
        {
            Unreal::FProperty* seat_side_prop = get_attached_seat_component_result.SeatComponent->GetPropertyByNameInChain(STR("SeatSide"));
            uint8_t* seat_side_num = seat_side_prop->ContainerPtrToValuePtr<uint8_t>(get_attached_seat_component_result.SeatComponent);
            if (*seat_side_num > 0)
            {
                train_side = 1;
            }
        }

        /* regex pattern to match {SIDE} with optional front/back placeholders */
        /* captures: front placeholder (optional), back placeholder (optional) */
        std::wregex side_placeholder_regex(STR(R"(\{SIDE(:[^:]+)?(:[^:]+)?\})"));
        std::wsmatch side_placeholder_matches;
        if (std::regex_search(control_name, side_placeholder_matches, side_placeholder_regex))
        {
            RC::StringType::const_iterator placeholder_start = side_placeholder_matches[0].first;
            RC::StringType::const_iterator placeholder_end = side_placeholder_matches[0].second;

            RC::StringType front_value = STR("F");
            RC::StringType back_value = STR("B");

            if (side_placeholder_matches.length() >= 2)
            {
                front_value = side_placeholder_matches[1].str().substr(1);
            }

            if (side_placeholder_matches.length() >= 3)
            {
                back_value = side_placeholder_matches[2].str().substr(1);
            }

            /* determine which value to use based on train side */
            RC::StringType train_side_str = train_side == 0 ? front_value : back_value;
            control_name.replace(placeholder_start, placeholder_end, train_side_str);
        }

        /* if no {SIDE} -> just return raw*/
        return control_name;
    }

    static Unreal::FName* get_vhid_component_input_identifier(Unreal::UObject* vhid_component)
    {
        Unreal::FStructProperty* input_identifier_prop =
                static_cast<Unreal::FStructProperty*>(vhid_component->GetPropertyByNameInChain(STR("InputIdentifier")));
        if (!input_identifier_prop) return nullptr;
        Unreal::UScriptStruct* input_identifier_struct = input_identifier_prop->GetStruct();
        auto input_identifier = input_identifier_prop->ContainerPtrToValuePtr<void>(vhid_component);
        Unreal::FProperty* input_identifier_identifier_prop = input_identifier_struct->GetPropertyByNameInChain(STR("Identifier"));
        return input_identifier_identifier_prop->ContainerPtrToValuePtr<Unreal::FName>(input_identifier);
    }

    static RC::StringType find_property_name_from_context(Unreal::UObject* actor, Unreal::UObject* context)
    {
        auto actor_class = actor->GetClassPrivate();
        for (Unreal::FProperty* prop = actor_class->GetPropertyLink(); prop; prop = prop->GetPropertyLinkNext())
        {
            auto prop_name = prop->GetName();
            if (Unreal::FObjectProperty* as_obj_prop = CastField<Unreal::FObjectProperty>(prop))
            {
                auto prop_value_ptr = as_obj_prop->ContainerPtrToValuePtr<void>(actor);
                if (as_obj_prop->GetPropertyValue(prop_value_ptr) == context)
                {
                    return prop_name;
                }
            }
        }
        return RC::StringType(STR(""));
    }

    static bool is_vhid_component_changing(Unreal::UObject* vhid_component)
    {
        if (!vhid_component) return false;

        Unreal::UFunction* is_changing_func = vhid_component->GetFunctionByNameInChain(STR("IsChanging"));
        if (!is_changing_func) return false;

        VirtualHIDComponent_IsChangingParams params;
        vhid_component->ProcessEvent(is_changing_func, &params);
        return params.IsChanging;
    }

    static float get_current_vhid_component_input_value(Unreal::UObject* vhid_component)
    {
        if (!vhid_component) return 0.0f;

        Unreal::UFunction* get_current_input_value_func = vhid_component->GetFunctionByNameInChain(STR("GetCurrentInputValue"));
        if (!get_current_input_value_func) return false;

        VirtualHIDComponent_GetCurrentInputValueParams params;
        vhid_component->ProcessEvent(get_current_input_value_func, &params);
        return params.InputValue;
    }

    static float get_current_vhid_component_normalized_input_value(Unreal::UObject* vhid_component)
    {
        if (!vhid_component) return 0.0f;

        Unreal::UFunction* get_normalised_input_value_func = vhid_component->GetFunctionByNameInChain(STR("GetNormalisedInputValue"));
        if (!get_normalised_input_value_func) return false;

        VirtualVHIDComponent_GetNormalisedInputValueParams params;
        vhid_component->ProcessEvent(get_normalised_input_value_func, &params);
        return params.InputValue;
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
        if (!TSWControllerMod::is_player_controller(controller)) return;

        // get pawn
        Unreal::UObject* pawn = TSWControllerMod::get_driver_pawn_from_controller(controller);
        Unreal::UFunction* get_drivable_actor_fn = controller->GetFunctionByNameInChain(STR("GetDrivableActor"));
        if (!pawn || !get_drivable_actor_fn) {
            Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Missing driver pawn or GetDrivableActor function\n"));
            return;
        }
        DriverController_GetDrivableActorParams drivable_actor_result;
        controller->ProcessEvent(get_drivable_actor_fn, &drivable_actor_result);
        if (!drivable_actor_result.DrivableActor) {
            return;
        }
        Unreal::UFunction* find_virtual_hid_component_func = drivable_actor_result.DrivableActor->GetFunctionByNameInChain(STR("FindVirtualHIDComponent"));
        Unreal::UFunction* notify_begin_interaction_func = controller->GetFunctionByNameInChain(STR("NotifyBeginInteraction"));
        Unreal::UFunction* begin_changing_vhid_component_func = controller->GetFunctionByNameInChain(STR("BeginChangingVHIDComponent"));

        /*
          used on the M3 MTA variant - if this is not called after updating controls; the constraints won't register properly
          this is not a problem on all trains
        */
        Unreal::UFunction* call_update_functions_func = drivable_actor_result.DrivableActor->GetFunctionByNameInChain(STR("CallUpdateFunctions"));

        if (!find_virtual_hid_component_func || !notify_begin_interaction_func || !begin_changing_vhid_component_func) return;

        std::unique_lock<std::shared_mutex> current_drivable_actor_lock(TSWControllerMod::CURRENT_DRIVABLE_ACTOR_CLASS_NAME_MUTEX);
        TSWControllerMod::TIME_SINCE_CURRENT_DRIVABLE_ACTOR_REPORTED += delta_secs;
        auto drivable_actor_name = drivable_actor_result.DrivableActor->GetClassPrivate()->GetName();
        if (
            TSWControllerMod::CURRENT_DRIVABLE_ACTOR_CLASS_NAME != drivable_actor_name ||
            TSWControllerMod::TIME_SINCE_CURRENT_DRIVABLE_ACTOR_REPORTED > 2.0f
        ) {
            TSWControllerMod::TIME_SINCE_CURRENT_DRIVABLE_ACTOR_REPORTED = 0.0f;
            TSWControllerMod::CURRENT_DRIVABLE_ACTOR_CLASS_NAME = drivable_actor_name;
            auto message = STR("current_drivable_actor,name=") + drivable_actor_name;
            auto message_str = std::string(message.begin(), message.end());
            Output::send<LogLevel::Default>(STR("[TSWControllerMod] sending current drivable actor information {}\n"), message);
            tsw_controller_mod_send_message((char*)message_str.c_str());
        }
        current_drivable_actor_lock.unlock();

        std::unique_lock<std::shared_mutex> direct_control_target_state_lock(TSWControllerMod::DIRECT_CONTROL_TARGET_STATE_MUTEX);
        std::unique_lock<std::shared_mutex> vhid_components_to_release_lock(TSWControllerMod::VHID_COMPONENTS_CHANGING_MUTEX);

        if (!TSWControllerMod::VHID_COMPONENTS_CHANGING.empty())
        {
            Unreal::UFunction* notify_end_interaction_func = controller->GetFunctionByNameInChain(STR("NotifyEndInteraction"));
            Unreal::UFunction* end_using_vhid_component_func = controller->GetFunctionByNameInChain(STR("EndUsingVHIDComponent"));
            if (!notify_end_interaction_func || !end_using_vhid_component_func) return;

            for (auto it = TSWControllerMod::VHID_COMPONENTS_CHANGING.begin(); it != TSWControllerMod::VHID_COMPONENTS_CHANGING.end();)
            {
                /* only release once there is no target state left */
                if (TSWControllerMod::DIRECT_CONTROL_TARGET_STATE.find(it->first) == TSWControllerMod::DIRECT_CONTROL_TARGET_STATE.end())
                {
                    Unreal::UObject* vhid_component = it->second.Get();
                    if (vhid_component)
                    {
                        PlayerController_EndUsingVHIDComponentParams params{vhid_component};
                        controller->ProcessEvent(end_using_vhid_component_func, &params);
                        controller->ProcessEvent(notify_end_interaction_func, &params);
                        Unreal::UFunction* end_changing_func = vhid_component->GetFunctionByNameInChain(STR("EndChanging"));
                        if (end_changing_func)
                        {
                            VirtualVHIDComponent_EndChangingParams params{controller};
                            vhid_component->ProcessEvent(end_changing_func, &params);
                        }
                        Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] stopped using VHID component: {}\n"), it->first);
                    }
                    it = TSWControllerMod::VHID_COMPONENTS_CHANGING.erase(it);
                }
                else
                {
                    ++it;
                }
            }
        }

        for (const auto& control_pair : TSWControllerMod::DIRECT_CONTROL_TARGET_STATE)
        {
            RC::StringType control_name = TSWControllerMod::format_direct_control_name(pawn, control_pair.first);
            RailVehicle_FindVirtualHIDComponentParams find_virtualhid_component_params = {Unreal::FName(control_name), nullptr};
            drivable_actor_result.DrivableActor->ProcessEvent(find_virtual_hid_component_func, &find_virtualhid_component_params);
            if (!find_virtualhid_component_params.VirtualHIDComponent)
            {
                continue;
            }

            Unreal::UFunction* set_pushed_state_func = find_virtualhid_component_params.VirtualHIDComponent->GetFunctionByNameInChain(STR("SetPushedState"));
            Unreal::UFunction* set_current_input_value_fn =
                    find_virtualhid_component_params.VirtualHIDComponent->GetFunctionByNameInChain(STR("SetCurrentInputValue"));
            Unreal::UFunction* set_normlised_input_value_fn =
                    find_virtualhid_component_params.VirtualHIDComponent->GetFunctionByNameInChain(STR("SetNormalisedInputValue"));

            auto target_value = control_pair.second.TargetValue;
            auto max_change_rate = control_pair.second.MaxChangeRate;
            auto flags = control_pair.second.Flags;
            bool should_hold = std::find(flags.begin(), flags.end(), STR("hold")) != flags.end();
            bool should_be_relative = std::find(flags.begin(), flags.end(), STR("relative")) != flags.end();
            bool should_use_normalized = std::find(flags.begin(), flags.end(), STR("normalized")) != flags.end();
            bool should_notify = std::find(flags.begin(), flags.end(), STR("notify")) != flags.end();
            auto get_current_value_func = should_use_normalized ? TSWControllerMod::get_current_vhid_component_normalized_input_value :  TSWControllerMod::get_current_vhid_component_input_value;
            auto set_input_value_func = should_use_normalized ? set_normlised_input_value_fn : set_current_input_value_fn;
            auto current_input_value = get_current_value_func(find_virtualhid_component_params.VirtualHIDComponent);

            /* account for relative flag */
            if (should_be_relative)
            {
                target_value = current_input_value + target_value;
                /* can't currently be used with hold */
                should_hold = false;
            }

            bool is_already_changing = TSWControllerMod::VHID_COMPONENTS_CHANGING.find(control_name) != TSWControllerMod::VHID_COMPONENTS_CHANGING.end();

            if (!is_already_changing)
            {
                PlayerController_BeginChangingVHIDComponentParams begin_changing_params{find_virtualhid_component_params.VirtualHIDComponent};
                controller->ProcessEvent(begin_changing_vhid_component_func, &begin_changing_params);
                TSWControllerMod::VHID_COMPONENTS_CHANGING[control_name] = Unreal::TWeakObjectPtr<Unreal::UObject>(find_virtualhid_component_params.VirtualHIDComponent);
                Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] started changing VHID component: {}\n"), control_name);

                if (should_notify)
                {
                    Controller_NotifyBeginInteractionParams notify_interaction_params{find_virtualhid_component_params.VirtualHIDComponent};
                    controller->ProcessEvent(notify_begin_interaction_func, &notify_interaction_params);
                }
            }

            /* apply incoming value */
            if (set_pushed_state_func)
            {
                VirtualHIDComponent_SetPushedStateParams set_pushed_state_params = {target_value > 0.5f, false};
                find_virtualhid_component_params.VirtualHIDComponent->ProcessEvent(set_pushed_state_func, &set_pushed_state_params);
                /* remove value from target states */
                if (!should_hold)
                {
                    TSWControllerMod::DIRECT_CONTROL_TARGET_STATE.erase(control_pair.first);
                }
            }
            else if (set_input_value_func)
            {
                /* check for max change rate */
                auto target_value_abs_diff = std::abs(current_input_value - target_value);
                auto target_set_value = current_input_value + (
                    current_input_value < target_value ? std::min(target_value_abs_diff, max_change_rate) : -std::min(target_value_abs_diff, max_change_rate)
                );

                VirtualHIDComponent_SetCurrentInputValueParams set_current_input_value_params = {target_set_value};
                find_virtualhid_component_params.VirtualHIDComponent->ProcessEvent(set_input_value_func, &set_current_input_value_params);
                /* check if value was reached within margin of error*/
                auto current_input_value = get_current_value_func(find_virtualhid_component_params.VirtualHIDComponent);
                if (!should_hold && TSWControllerMod::is_within_margin_of_error(target_value, current_input_value))
                {
                    Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] clearing target value for control {}\n"), control_pair.first);
                    /* remove value from target states */
                    TSWControllerMod::DIRECT_CONTROL_TARGET_STATE.erase(control_pair.first);
                }
            }

            // /* run post update functions */
            // if (call_update_functions_func)
            // {
            //     Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Calling M3 MTA specific CallUpdateFunctions\n"));
            //     EmptyVoidFunctionParams params{};
            //     drivable_actor_result.DrivableActor->ProcessEvent(call_update_functions_func, &params);
            // }
        }

        vhid_components_to_release_lock.unlock();
        direct_control_target_state_lock.unlock();
    }

    static void on_direct_control_message_received(const char* raw_message)
    {
        /* update DC target state */
        std::unique_lock<std::shared_mutex> lock(TSWControllerMod::DIRECT_CONTROL_TARGET_STATE_MUTEX);

        auto message = RC::ensure_str(std::string{raw_message});
        auto parts = TSWControllerMod::wstring_split(message, STR(","));
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

        Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Processing Direct Control message: {}\n"), message);
        float target_value = std::stof(properties[STR("value")]);
        float max_change_rate = (properties.find(STR("max_change_rate")) == properties.end()) ? 1000.0 : std::stof(properties[STR("max_change_rate")]);
        std::vector<RC::StringType> flags = TSWControllerMod::wstring_split(properties[STR("flags")], STR("|"));
        DirectControlControlTargetState control_target_state{target_value,max_change_rate,flags};
        TSWControllerMod::DIRECT_CONTROL_TARGET_STATE[properties[STR("controls")]] = control_target_state;
    }

    static void on_ts2_virtualhidcomponent_inputvaluechanged(Unreal::UnrealScriptFunctionCallableContext context, void* custom_data)
    {
        Unreal::FName* input_identifier = TSWControllerMod::get_vhid_component_input_identifier(context.Context);
        Unreal::UFunction* get_currently_changing_controller_func = context.Context->GetFunctionByNameInChain(STR("GetCurrentlyChangingController"));
        if (input_identifier && get_currently_changing_controller_func)
        {
            VirtualHIDComponent_GetCurrentlyChangingControllerParams get_currently_changing_controller_params{};
            context.Context->ProcessEvent(get_currently_changing_controller_func, &get_currently_changing_controller_params);
            /* don't do anything if it's a none identifier, there is no controller or it's not the player controller */
            if (!get_currently_changing_controller_params.Controller || !TSWControllerMod::is_player_controller(get_currently_changing_controller_params.Controller))
            {
                return;
            }

            /* find drivable actor*/
            Unreal::UFunction* get_drivable_actor_fn = get_currently_changing_controller_params.Controller->GetFunctionByNameInChain(STR("GetDrivableActor"));
            if (!get_drivable_actor_fn) {
                Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Can't find GetDrivableActor function\n"));
                return;
            }
            DriverController_GetDrivableActorParams drivable_actor_result;
            get_currently_changing_controller_params.Controller->ProcessEvent(get_drivable_actor_fn, &drivable_actor_result);
            if (!drivable_actor_result.DrivableActor) {
                return;
            }

            /* loop over class properties to find raw controller identifier */
            auto control_property_name = TSWControllerMod::find_property_name_from_context(drivable_actor_result.DrivableActor, context.Context);

            /* if we can't find a property it's likely not relevant so we can ignore */
            if (control_property_name.empty())
            {
                return;
            }

            /* get normalised input value */
            auto normalized_value = TSWControllerMod::get_current_vhid_component_normalized_input_value(context.Context);

            /* send updated value */
            VirtualHIDComponent_InputValueChangedParams input_value_changed_params = context.GetParams<VirtualHIDComponent_InputValueChangedParams>();
            /* message format = sync_control,name={name},property={control_property_name},value={value},normal_value={normal_value} */
            auto message = STR("sync_control_value,name=") + input_identifier->ToString() + STR(",property=") + control_property_name + STR(",value=") + std::to_wstring(input_value_changed_params.NewValue) + STR(",normalized_value=") + std::to_wstring(normalized_value);
            auto message_str = std::string(message.begin(), message.end());
            Output::send<LogLevel::Default>(STR("[TSWControllerMod] sending updated control value {}\n"), message);
            tsw_controller_mod_send_message((char*)message_str.c_str());
        }
    }

  public:
    TSWControllerMod() : CppUserModBase()
    {
        ModName = STR("TSWControllerMod");
        ModVersion = STR("1.0.0");
        ModDescription = STR("TSW Controller Utility Helper");
        ModAuthors = STR("Liam");

        Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Starting..."));
    }

    auto on_unreal_init() -> void override
    {
        Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Unreal Initialized"));

        Unreal::UFunction* input_value_changed_func =
                Unreal::UObjectGlobals::StaticFindObject<Unreal::UFunction*>(nullptr, nullptr, STR("/Script/TS2Prototype.VirtualHIDComponent:InputValueChanged"));
        if (!input_value_changed_func) return;

        Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Registering hooks and callbacks"));
        input_value_changed_func->RegisterPostHook(TSWControllerMod::on_ts2_virtualhidcomponent_inputvaluechanged);
        Unreal::Hook::RegisterAActorTickPreCallback(TSWControllerMod::on_tick);
        tsw_controller_mod_set_receive_message_callback(TSWControllerMod::on_direct_control_message_received);
    }

    ~TSWControllerMod() override = default;
};

#define TSW_CONTROLLER_MOD_API __declspec(dllexport)
extern "C"
{
    TSW_CONTROLLER_MOD_API RC::CppUserModBase* start_mod()
    {
        tsw_controller_mod_start();
        return new TSWControllerMod();
    }

    TSW_CONTROLLER_MOD_API void uninstall_mod(RC::CppUserModBase* mod)
    {
        tsw_controller_mod_stop();
        delete mod;
    }
}
