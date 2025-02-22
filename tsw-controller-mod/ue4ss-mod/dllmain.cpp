#include <string>
#include <format>

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
#include <DynamicOutput/Output.hpp>
#include <UE4SSProgram.hpp>

#include "tsw_controller_mod_socket_connection.h"

class TSWControllerMod : public RC::CppUserModBase
{
private:
  static inline Unreal::UObject *PLAYER_CONTROLLER = nullptr;

  static Unreal::UObject *get_player_controller()
  {
    if (TSWControllerMod::PLAYER_CONTROLLER)
    {
      return TSWControllerMod::PLAYER_CONTROLLER;
    }

    std::vector<Unreal::UObject *> player_controllers{};
    Unreal::UObjectGlobals::FindAllOf(STR("PlayerController"), player_controllers);
    if (!player_controllers.empty())
    {
      TSWControllerMod::PLAYER_CONTROLLER = player_controllers.back();
      return TSWControllerMod::PLAYER_CONTROLLER;
    }
    return nullptr;
  }

  static Unreal::FName *get_vhid_component_input_identifier(Unreal::UObject *vhid_component)
  {
    Unreal::FStructProperty *input_identifier_prop =
        static_cast<Unreal::FStructProperty *>(vhid_component->GetPropertyByNameInChain(STR("InputIdentifier")));
    if (!input_identifier_prop)
      return nullptr;
    Unreal::UScriptStruct *input_identifier_struct = input_identifier_prop->GetStruct();
    auto input_identifier = input_identifier_prop->ContainerPtrToValuePtr<void>(vhid_component);
    Unreal::FProperty *input_identifier_identifier_prop = input_identifier_struct->GetPropertyByNameInChain(STR("Identifier"));
    return input_identifier_identifier_prop->ContainerPtrToValuePtr<Unreal::FName>(input_identifier);
  }

  static void on_direct_control_message_received(const char* message)
  {
    /* @TODO apply to loco*/
  }

  static void on_ts2_virtualhidcomponent_inputvaluechanged(Unreal::UnrealScriptFunctionCallableContext context, void *custom_data)
  {
    Unreal::UObject *player_controller = TSWControllerMod::get_player_controller();
    Unreal::FName *input_identifier = TSWControllerMod::get_vhid_component_input_identifier(context.Context);
    Unreal::UFunction *get_currently_changing_controller_func = context.Context->GetFunctionByNameInChain(STR("GetCurrentlyChangingController"));
    if (input_identifier && get_currently_changing_controller_func)
    {

      struct Params
      {
        Unreal::UObject *Controller;
      };
      Params params{};
      context.Context->ProcessEvent(get_currently_changing_controller_func, &params);

      if (input_identifier->ToString() != STR("None") && params.Controller && params.Controller == player_controller)
      {
        auto message = std::format("{},{}", input_identifier->ToString(), "value");
        tsw_controller_mod_send_sync_controller_message(message.c_str());
        Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] SyncInputValue {}..."), message);
      }
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

    Unreal::UFunction *unreal_function =
        Unreal::UObjectGlobals::StaticFindObject<Unreal::UFunction *>(nullptr, nullptr, STR("/Script/TS2Prototype.VirtualHIDComponent:InputValueChanged"));
    if (!unreal_function)
      return;

    auto func_ptr = unreal_function->GetFunc();
    if (!func_ptr)
      return;

    Output::send<LogLevel::Verbose>(STR("[TSWControllerMod] Hook Registered"));
    unreal_function->RegisterPostHook(on_ts2_virtualhidcomponent_inputvaluechanged);
  }

  ~TSWControllerMod() override = default;
};

#define TSW_CONTROLLER_MOD_API __declspec(dllexport)
extern "C"
{
  TSW_CONTROLLER_MOD_API RC::CppUserModBase *start_mod()
  {
    tsw_controller_mod_start();
    return new TSWControllerMod();
  }

  TSW_CONTROLLER_MOD_API void uninstall_mod(RC::CppUserModBase *mod)
  {
    /* @TODO stop listeners?*/
    delete mod;
  }
}