import { useField } from 'formik';
import { classNames } from '~/utils/helpers';

type InputProps = {
  id: string;
  name: string;
  type?: string;
  className?: string;
  autoComplete?: boolean;
  forwardRef?: React.RefObject<HTMLInputElement>;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
} & React.InputHTMLAttributes<HTMLInputElement>;

export default function Input(props: InputProps) {
  const {
    id,
    type = 'text',
    className = '',
    onChange,
    autoComplete = false,
    forwardRef,
    ...rest
  } = props;

  const [field, meta] = useField(props);
  const hasError = !!(meta.touched && meta.error);

  return (
    <>
      <input
        ref={forwardRef}
        className={classNames(
          hasError ? 'border-red-400' : 'border-gray-300',
          `${className} block w-full rounded-md shadow-sm focus:ring-violet-300 focus:border-violet-300 disabled:cursor-not-allowed disabled:text-gray-500 disabled:bg-gray-50 disabled:border-gray-200 sm:text-sm`
        )}
        id={id}
        type={type}
        {...field}
        onChange={(e) => {
          field.onChange(e);
          onChange && onChange(e);
        }}
        autoComplete={autoComplete ? 'on' : 'off'}
        {...rest}
      />
      {meta.touched && meta.error ? (
        <div className="mt-1 text-sm text-red-500">{meta.error}</div>
      ) : null}
    </>
  );
}
