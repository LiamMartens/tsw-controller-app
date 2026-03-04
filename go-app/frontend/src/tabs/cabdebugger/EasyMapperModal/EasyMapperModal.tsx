import z from "zod";
import { Modal, ModalContentProps } from "../../../components";
import { useEffect, useMemo } from "react";
import { useCabControlState, useControllers } from "../../../swr";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEasyMapperModalForm } from "./useEasyMapperModalForm";
import { EasyMapperModalFormControlRow } from "./EasyMapperModalFormControlRow";

type Props = {
  open: boolean;
  onClose: () => void;
};

const EasyMapperModalContent = ({ onClose }: ModalContentProps<true>) => {
  const { data: controllers } = useControllers();
  const { data: cabControlState } = useCabControlState({
    refreshInterval: 100,
  });

  const form = useEasyMapperModalForm();

  const selectedControllerUniqueID = form.watch("controller");
  const controls = form.watch("controls");

  const selectableControllers = useMemo(
    () => controllers.filter((c) => !c.IsVirtual && c.IsConfigured),
    [controllers],
  );

  const selectedController = useMemo(
    () =>
      selectableControllers.find(
        (c) => c.UniqueID === selectedControllerUniqueID,
      ),
    [selectableControllers, selectedControllerUniqueID],
  );

  const handleAddControl = () => {
    form.setValue("controls", [
      ...form.getValues("controls"),
      { type: "button", control: "", binding: { name: "" } },
    ]);
  };

  return (
    <form className="flex flex-col gap-3">
      <fieldset className="fieldset">
        <legend className="fieldset-legend">Select controller</legend>
        <select
          disabled={!!selectedController}
          className="select w-full"
          {...form.register("controller")}
        >
          <option disabled value="">
            Select controller
          </option>
          {selectableControllers.map((c) => (
            <option key={c.UniqueID} value={c.UniqueID}>
              {c.Name} ({c.DeviceID})
            </option>
          ))}
        </select>
        <span className="label">Only configured controllers are shown</span>
      </fieldset>

      {!!selectedController && (
        <div className="flex flex-col gap-3">
          {controls.map((control, index) => (
            <EasyMapperModalFormControlRow
              key={`control_${index}`}
              cabControlState={cabControlState}
              form={form}
              index={index}
            />
          ))}
          <button
            type="button"
            className="btn btn-sm w-full"
            onClick={handleAddControl}
          >
            Add mapping
          </button>
        </div>
      )}
    </form>
  );
};

export const EasyMapperModal = ({ open, onClose }: Props) => {
  return (
    <Modal
      openState={open}
      onClose={onClose}
      Component={EasyMapperModalContent}
    />
  );
};
