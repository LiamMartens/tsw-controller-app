#include <string>
#include <format>
#include <mutex>
#include <queue>
#include <shared_mutex>
#include <unordered_map>

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

struct VirtualHIDComponent_GetCurrentlyChangingControllerParams { Unreal::UObject* Controller; };
struct VirtualHIDComponent_InputValueChangedParams { float OldValue; float NewValue; };
struct PlayerController_IsPlayerControllerParams { bool IsPlayerController; };
struct PlayerController_GetDriverPawnParams { Unreal::UObject* DriverPawn; };
struct DriverPawn_GetAttachedSeatComponentParams { Unreal::UObject* SeatComponent; };
struct DriverPawn_GetDrivableActorParams { Unreal::UObject* DrivableActor; };
struct RailVehicle_FindVirtualHIDComponentParams { Unreal::FName Name; Unreal::UObject* VirtualHIDComponent; };
struct VirtualHIDComponent_SetCurrentInputValueParams { float Value; };
struct VirtualHIDComponent_SetPushedStateParams { bool IsPushed; bool SkipTransition; };
struct PlayerController_BeginChangingVHIDComponentParams { Unreal::UObject* Component; };
struct PlayerController_EndUsingVHIDComponentParams { Unreal::UObject* Component; };
struct GameplayStatistics_GetPlayerControllerParams { Unreal::UWorld* World; int32_t PlayerIndex; Unreal::UObject* PlayerController; };

class TSWControllerMod : public RC::CppUserModBase
{
private:
  static inline std::shared_mutex DIRECT_CONTROL_QUEUE_MUTEX;
  static inline std::queue<RC::StringType> DIRECT_CONTROL_QUEUE;

  static bool is_player_controller(Unreal::UObject* controller)
  {
    PlayerController_IsPlayerControllerParams is_player_controller_result;
    Unreal::UFunction* is_player_function = controller->GetFunctionByNameInChain(STR("IsPlayerController"));
    if (is_player_function)
    {
      controller->ProcessEvent(is_player_function, &is_player_controller_result);
      return is_player_controller_result.IsPlayerController;
    }
    return false;
  }

  static Unreal::UObject* get_player_controller_from(Unreal::UObject* from)
  {
    Unreal::UGameplayStatics* statistics = Unreal::UObjectGlobals::StaticFindObject<Unreal::UGameplayStatics*>(nullptr, nullptr, STR("/Script/Engine.Default__GameplayStatics"));
    Unreal::UFunction* get_player_controller_func = statistics ? statistics->GetFunctionByNameInChain(STR("GetPlayerController")) : nullptr;
    if (from && statistics && get_player_controller_func)
    {
      GameplayStatistics_GetPlayerControllerParams params{ from->GetWorld() };
      statistics->ProcessEvent(get_player_controller_func, &params);
      if (params.PlayerController)
      {
        return params.PlayerController;
      }
    }

    std::vector<Unreal::UObject*> player_controllers{};
    Unreal::UObjectGlobals::FindAllOf(STR("PlayerController"), player_controllers);
    for (Unreal::UObject* controller : player_controllers)
    {
      if (TSWControllerMod::is_player_controller(controller))
      {
        return controller;
      }
    }
    return nullptr;
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
      if (*seat_side_num == 1)
      {
        train_side = 1;
      }
    }

    RC::StringType train_side_placeholder = STR("{SIDE}");
    std::size_t side_placeholder_pos = control_name.find(train_side_placeholder);
    /* if no {SIDE} -> just return raw*/
    if (side_placeholder_pos != RC::StringType::npos)
    {
      RC::StringType train_side_str = train_side == 0 ? STR("F") : STR("B");
      control_name.replace(side_placeholder_pos, train_side_placeholder.length(), train_side_str);
    }
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

  static void on_process_event_pre_callback(Unreal::UObject* context, Unreal::UFunction* function, void* params)
  {
    context->GetWorld();
    /* this is functional but slow when called too many times */

    if (function->GetName() != STR("Tick"))
    {
      /* only run on ticks to prevent clogging*/
      return;
    }

    std::shared_lock<std::shared_mutex> lock(TSWControllerMod::DIRECT_CONTROL_QUEUE_MUTEX);
    if (TSWControllerMod::DIRECT_CONTROL_QUEUE.empty()) return;

    /* aggregate all dc messages if more than one*/
    std::unordered_map<RC::StringType, float> target_input_values;
    while (!TSWControllerMod::DIRECT_CONTROL_QUEUE.empty())
    {
      RC::StringType message = TSWControllerMod::DIRECT_CONTROL_QUEUE.front();
      TSWControllerMod::DIRECT_CONTROL_QUEUE.pop();
      auto parts = TSWControllerMod::wstring_split(message, STR(","));
      if (parts.size() < 3 || parts[0] != STR("direct_control")) continue;
      Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Processing Direct Control message: {}\n"), message);
      target_input_values[parts[1]] = std::stof(parts[2]);
    }

    if (target_input_values.empty()) return;

    /* skip if no controller or pawn */
    Unreal::UObject* controller = TSWControllerMod::get_player_controller_from(context);
    Unreal::UObject* pawn = TSWControllerMod::get_driver_pawn_from_controller(controller);
    if (!controller || !pawn) return;

    /* skip if drivable actor can't be found */
    Unreal::UFunction* get_drivable_actor_fn = pawn->GetFunctionByNameInChain(STR("GetDrivableActor"));
    if (!get_drivable_actor_fn) return;

    DriverPawn_GetDrivableActorParams drivable_actor_result;
    pawn->ProcessEvent(get_drivable_actor_fn, &drivable_actor_result);
    if (!drivable_actor_result.DrivableActor) return;

    Unreal::UFunction* find_virtual_hid_component_func = drivable_actor_result.DrivableActor->GetFunctionByNameInChain(STR("FindVirtualHIDComponent"));
    if (!find_virtual_hid_component_func) return;

    for (const auto& control_pair : target_input_values)
    {
      RC::StringType control_name = TSWControllerMod::format_direct_control_name(pawn, control_pair.first);
      RailVehicle_FindVirtualHIDComponentParams find_virtualhid_component_params = { Unreal::FName(control_name), nullptr };
      drivable_actor_result.DrivableActor->ProcessEvent(find_virtual_hid_component_func, &find_virtualhid_component_params);
      if (!find_virtualhid_component_params.VirtualHIDComponent) continue;

      /* @TODO begin changing*/
      Unreal::UFunction* begin_changing_func = controller->GetFunctionByNameInChain(STR("BeginChangingVHIDComponent"));
      Unreal::UFunction* end_using_func = controller->GetFunctionByNameInChain(STR("EndUsingVHIDComponent"));
      Unreal::UFunction* set_pushed_state_func = find_virtualhid_component_params.VirtualHIDComponent->GetFunctionByNameInChain(STR("SetPushedState"));
      Unreal::UFunction* set_current_input_value_fn = find_virtualhid_component_params.VirtualHIDComponent->GetFunctionByNameInChain(STR("SetCurrentInputValue"));
      if (begin_changing_func)
      {
        PlayerController_BeginChangingVHIDComponentParams params{ find_virtualhid_component_params.VirtualHIDComponent };
        controller->ProcessEvent(begin_changing_func, &params);
      }
      if (set_pushed_state_func)
      {
        VirtualHIDComponent_SetPushedStateParams set_pushed_state_params = { control_pair.second > 0.5f, true };
        find_virtualhid_component_params.VirtualHIDComponent->ProcessEvent(set_pushed_state_func, &set_pushed_state_params);
      }
      else if (set_current_input_value_fn)
      {
        VirtualHIDComponent_SetCurrentInputValueParams set_current_input_value_params = { control_pair.second };
        find_virtualhid_component_params.VirtualHIDComponent->ProcessEvent(set_current_input_value_fn, &set_current_input_value_params);
      }
      if (end_using_func)
      {
        PlayerController_EndUsingVHIDComponentParams params{ find_virtualhid_component_params.VirtualHIDComponent };
        controller->ProcessEvent(end_using_func, &params);
      }
    }
  }

  static void on_direct_control_message_received(const char* raw_message)
  {
    /* push DC message into the back */
    std::unique_lock<std::shared_mutex> lock(TSWControllerMod::DIRECT_CONTROL_QUEUE_MUTEX);
    auto message = RC::ensure_str(std::string{ raw_message });
    TSWControllerMod::DIRECT_CONTROL_QUEUE.push(message);
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
      if (
        input_identifier->ToString() == STR("None")
        || !get_currently_changing_controller_params.Controller
        || !TSWControllerMod::is_player_controller(get_currently_changing_controller_params.Controller)
        ) {
        return;
      }

      VirtualHIDComponent_InputValueChangedParams inptu_value_changed_params = context.GetParams<VirtualHIDComponent_InputValueChangedParams>();
      auto message = input_identifier->ToString() + STR(",") + std::to_wstring(inptu_value_changed_params.NewValue);
      Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Sending SC message {}\n"), message);
      tsw_controller_mod_send_sync_controller_message(std::string(message.begin(), message.end()).c_str());
    }
  }

public:
  TSWControllerMod() : CppUserModBase()
  {
    ModName = STR("TSWControllerMod");
    ModVersion = STR("1.0");
    ModDescription = STR("TSW Direct Access Controller");
    ModAuthors = STR("truman");

    Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Starting..."));
  }

  auto on_unreal_init() -> void override
  {
    Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Unreal Initialized"));

    Unreal::UFunction* unreal_function =
      Unreal::UObjectGlobals::StaticFindObject<Unreal::UFunction*>(nullptr, nullptr, STR("/Script/TS2Prototype.VirtualHIDComponent:InputValueChanged"));
    if (!unreal_function) return;

    auto func_ptr = unreal_function->GetFunc();
    if (!func_ptr) return;

    Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Registering hooks and callbacks"));
    Unreal::Hook::RegisterProcessEventPreCallback(TSWControllerMod::on_process_event_pre_callback);
    unreal_function->RegisterPostHook(TSWControllerMod::on_ts2_virtualhidcomponent_inputvaluechanged);
    tsw_controller_mod_set_direct_controller_callback(TSWControllerMod::on_direct_control_message_received);
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
    /* @TODO stop listeners?*/
    delete mod;
  }
}