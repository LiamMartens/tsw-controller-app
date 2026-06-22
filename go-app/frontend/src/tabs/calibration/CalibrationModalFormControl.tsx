import "react-bezier-curve-editor/index.css";
import { easings } from "animejs";
import { BezierCurveEditor, ValueType } from "react-bezier-curve-editor";
import {
  CalibrationStateControl,
  UseCalibrationFormType,
} from "./useCalibrationForm";
import { CSSProperties, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Controller } from "react-hook-form";
import { Modal, ModalContentProps } from "../../components";

type Props = {
  form: UseCalibrationFormType;
  index: number;
  field: CalibrationStateControl;
  isRunning: boolean;
};

const CalibrationModalFormControlEditCurveModal = ({
  openState: { index, field, form },
}: ModalContentProps<Props>) => {
  return (
    <div>
      <form method="dialog">
        <button className="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">
          ✕
        </button>
      </form>
      <div className="flex flex-col gap-4">
        <h3 className="font-bold text-sm">Editing {field.name} curve</h3>
        <div
          style={
            {
              "--bce-colors-background": "var(--color-base-300)",
              "--bce-colors-row": "var(--color-base-100)",
              "--bce-colors-outerarea": "var(--color-base-100)",
              "--bce-colors-handle-fixed": "var(--color-base-content)",
              "--bce-colors-handle-start": "var(--color-secondary)",
              "--bce-colors-handle-end": "var(--color-secondary)",
              "--bce-colors-curve-line": "var(--color-primary)",
              "--bce-colors-preview": "var(--color-base-content)",
              "--bce-colors-preview-border": "var(--color-primary)",
            } as CSSProperties
          }
          className="flex flex-col gap-2"
        >
          <Controller
            control={form.control}
            name={`controls.${index}.easingCurve`}
            render={({ field }) => (
              <div className="flex justify-center">
                <BezierCurveEditor
                  enablePreview
                  size={255}
                  outerAreaSize={0}
                  value={field.value as ValueType}
                  onChange={field.onChange}
                />
              </div>
            )}
          />
          <p className="text-xs text-base-content/60">
            The axis curve controls how your physical axis behaves. By default
            the axis is linear, but this can be adjusted to make the behavior
            faster or slower depending on the curve.
          </p>
        </div>
        <form method="dialog">
          <button className="btn btn-sm w-full">Close</button>
        </form>
      </div>
    </div>
  );
};

export const CalibrationModalFormControl = (props: Props) => {
  const { form, index, field, isRunning } = props;
  const [editCurveDialogOpen, setEditCurveDialogOpen] = useState(false);
  const normalAxisValue = useMemo(() => {
    const ease = easings.cubicBezier(...field.easingCurve);

    if (!field.invert) {
      return ease(
        (field.value + Math.abs(field.min)) / (Math.abs(field.min) + field.max),
      );
    }
    return ease(
      (Math.abs(field.min) + field.max - (field.value + Math.abs(field.min))) /
        (Math.abs(field.min) + field.max),
    );
  }, [field]);

  const handleEditAxisCurve = () => {
    setEditCurveDialogOpen(true);
  };

  return (
    <div
      key={`${field.kind}_${field.index}`}
      className="card card-sm shadow-sm"
    >
      <div className="card-body">
        <div>
          <div className="flex flex-col basis-full gap-2">
            <div className="flex justify-between items-center">
              <div>
                {field.kind} {field.index}
              </div>
              <div>
                <kbd className="kbd kbd-sm">{field.value}</kbd>
              </div>
            </div>
            {field.kind === "axis" && (
              <div>
                <progress
                  className="progress progress-primary w-full"
                  value={normalAxisValue}
                  max={1}
                ></progress>
              </div>
            )}
            <div className="grid grid-cols-2 grid-flow-row auto-rows-max gap-2">
              <label className="input input-xs w-full">
                Name
                <input
                  type="text"
                  className="grow"
                  disabled={!isRunning}
                  {...form.register(`controls.${index}.name`)}
                />
              </label>
            </div>
            {field.kind === "axis" && (
              <>
                <div className="grid grid-cols-2 grid-flow-row auto-rows-max gap-2">
                  <label className="input input-xs w-full">
                    Min
                    <input
                      type="number"
                      className="grow"
                      disabled={!field.override}
                      {...form.register(`controls.${index}.min`, {
                        valueAsNumber: true,
                        required: true,
                      })}
                    />
                  </label>
                  <label className="input input-xs w-full">
                    Max
                    <input
                      type="number"
                      className="grow"
                      disabled={!field.override}
                      {...form.register(`controls.${index}.max`, {
                        valueAsNumber: true,
                        required: true,
                      })}
                    />
                  </label>
                  <label className="input input-xs w-full">
                    Idle
                    <input
                      type="number"
                      className="grow"
                      disabled={!field.override}
                      {...form.register(`controls.${index}.idle`, {
                        valueAsNumber: true,
                        required: true,
                      })}
                    />
                  </label>
                  <label className="input input-xs w-full">
                    Deadzone
                    <input
                      type="number"
                      className="grow"
                      disabled={!field.override}
                      {...form.register(`controls.${index}.deadzone`, {
                        valueAsNumber: true,
                        required: true,
                      })}
                    />
                  </label>
                  <label className="input input-xs w-full">
                    Anti-Jitter
                    <input
                      type="number"
                      className="grow"
                      disabled={!field.override}
                      {...form.register(`controls.${index}.antiJitter`, {
                        valueAsNumber: true,
                        required: true,
                      })}
                    />
                  </label>
                </div>
                <div className="flex justify-start gap-2 items-center">
                  <div>
                    <label className="label">
                      <input
                        type="checkbox"
                        disabled={!isRunning}
                        className="checkbox checkbox-xs"
                        {...form.register(`controls.${index}.invert`)}
                      />
                      Invert
                    </label>
                  </div>
                  <div>
                    <label className="label">
                      <input
                        type="checkbox"
                        className="checkbox checkbox-xs"
                        disabled={!isRunning}
                        {...form.register(`controls.${index}.override`)}
                      />
                      Override calibration values
                    </label>
                  </div>
                  <button
                    disabled={!isRunning}
                    className="ml-auto btn btn-xs"
                    onClick={handleEditAxisCurve}
                  >
                    Edit axis curve
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>

      {createPortal(
        <Modal
          openState={editCurveDialogOpen ? props : false}
          onClose={() => setEditCurveDialogOpen(false)}
          Component={CalibrationModalFormControlEditCurveModal}
        />,
        document.body,
      )}
    </div>
  );
};
