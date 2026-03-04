import { main } from "../../../../wailsjs/go/models";
import { TUseEasyMapperModalFormReturn } from "./useEasyMapperModalForm";

type Props = {
  form: TUseEasyMapperModalFormReturn;
  index: number;
  cabControlState: main.Interop_Cab_ControlState;
};

export const EasyMapperModalFormControlRow = ({
  index,
  form,
  cabControlState,
}: Props) => {
  const controlType = form.watch(`controls.${index}.type`);

  return (
    <div className="flex flex-col gap-3">
      <div className="grid gap-2 grid-cols-[minmax(0,1fr)_minmax(0,1fr)_max-content]">
        <fieldset className="fieldset">
          <legend className="fieldset-legend">Select cab control</legend>
          <select
            className="select w-full"
            {...form.register(`controls.${index}.control`)}
          >
            <option disabled value="">
              Select cab control
            </option>
            {cabControlState.Controls.map((c) => (
              <option key={c.PropertyName} value={c.PropertyName}>
                {c.PropertyName}
              </option>
            ))}
          </select>
        </fieldset>

        <fieldset className="fieldset">
          <legend className="fieldset-legend">Select control type</legend>
          <select
            className="select w-full"
            {...form.register(`controls.${index}.type`)}
          >
            <option value="button">Button</option>
            <option value="switch/simple">Switch (On/Off)</option>
            <option value="switch/multi">Switch (Multi-Value)</option>
            <option value="lever">Lever</option>
          </select>
        </fieldset>

        {(controlType === "button" || controlType === "switch/simple") && (
          <button type="button" className="mt-auto mb-1 btn btn-md">Bind Control</button>
        )}
        {(controlType === "switch/multi" || controlType === "lever") && (
          <button type="button" className="mt-auto mb-1 btn btn-md">Add value</button>
        )}
      </div>
    </div>
  );
};
