import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useOutletContext } from 'react-router-dom';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import DeletePanel from '~/components/DeletePanel';
import FlagForm from '~/components/flags/FlagForm';
import VariantForm from '~/components/flags/VariantForm';
import Modal from '~/components/Modal';
import MoreInfo from '~/components/MoreInfo';
import Slideover from '~/components/Slideover';
import { deleteVariant } from '~/data/api';
import { FlagType, toFlagType } from '~/types/Flag';
import { IVariant } from '~/types/Variant';
import { FlagProps } from './FlagProps';
import Variants from './Variants';

export default function EditFlag() {
  const { flag, onFlagChange } = useOutletContext<FlagProps>();

  const [showVariantForm, setShowVariantForm] = useState<boolean>(false);
  const [editingVariant, setEditingVariant] = useState<IVariant | null>(null);
  const [showDeleteVariantModal, setShowDeleteVariantModal] =
    useState<boolean>(false);
  const [deletingVariant, setDeletingVariant] = useState<IVariant | null>(null);

  const variantFormRef = useRef(null);

  const namespace = useSelector(selectCurrentNamespace);

  return (
    <>
      {/* variant edit form */}
      <Slideover
        open={showVariantForm}
        setOpen={setShowVariantForm}
        ref={variantFormRef}
      >
        <VariantForm
          ref={variantFormRef}
          flagKey={flag.key}
          variant={editingVariant || undefined}
          setOpen={setShowVariantForm}
          onSuccess={() => {
            setShowVariantForm(false);
            onFlagChange();
          }}
        />
      </Slideover>

      {/* variant delete modal */}
      <Modal open={showDeleteVariantModal} setOpen={setShowDeleteVariantModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the variant{' '}
              <span className="font-medium text-violet-500">
                {deletingVariant?.key}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Variant"
          setOpen={setShowDeleteVariantModal}
          handleDelete={
            () =>
              deleteVariant(namespace.key, flag.key, deletingVariant?.id ?? '') // TODO: Determine impact of blank ID param
          }
          onSuccess={() => {
            onFlagChange();
          }}
        />
      </Modal>

      <div className="flex flex-col">
        {/* flag details */}
        <div className="my-10">
          <div className="md:grid md:grid-cols-3 md:gap-6">
            <div className="md:col-span-1">
              <p className="mt-1 text-sm text-gray-500">
                Basic information about the flag and its status.
              </p>
              <MoreInfo
                className="mt-5"
                href="https://www.flipt.io/docs/concepts#flags"
              >
                Learn more about flags
              </MoreInfo>
            </div>
            <div className="mt-5 md:col-span-2 md:mt-0">
              <FlagForm flag={flag} flagChanged={onFlagChange} />
            </div>
          </div>
        </div>

        {flag && toFlagType(flag.type) === FlagType.VARIANT_FLAG_TYPE && (
          <Variants flag={flag} onFlagChange={onFlagChange} />
        )}
      </div>
    </>
  );
}
