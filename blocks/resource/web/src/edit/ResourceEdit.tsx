"use client";

import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ComponentType,
  type ReactNode,
} from "react";
import { ResourceFieldKind, type ResourceEditState, type ResourceRelatedLink } from "./types";
import {
  buildSubmitValues,
  cloneResourceValues,
  formatSectionTitle,
  getResourceSubmitButtonState,
  groupFieldsBySection,
  isResourceDraftDirty,
  isResourceRecordNew,
} from "./resource-edit";
import {
  firstValidationError,
  mapSubmitErrorToFieldErrors,
  normalizeSubmitValues,
  validateDirtyFields,
  validateResourceValues,
} from "./resource-validate";
import type { ResourceEditMode, ResourceFormHandlers } from "./handlers";
import {
  resolveModeDescription,
  resolveResourceListAndEditCopy,
  type ResourceListAndEditDescriptions,
} from "./resource-descriptions";
import { ResourceFieldInput } from "./ResourceFieldInput";
import { ResourceRelatedLinks } from "./ResourceRelatedLinks";
import { geocodeAddressPreview } from "./geocode-preview";
import {
  buildAddressQueryFromDraft,
  buildPendingLocationValue,
  buildPreviewLocationValue,
  fieldTriggersLocationPreview,
  parseLocationValue,
  previewLocationFieldKeys,
} from "./resource-location";

const DEFAULT_LOCATION_PREVIEW_DELAY_MS = 600;

export type ResourceEditProps = {
  state: ResourceEditState;
  handlers: ResourceFormHandlers;
  saving?: boolean;
  twoColumnLayout?: boolean;
  /** Debounce for client-side map geocode preview (no server call). */
  locationPreviewDelayMs?: number;
  /** Resource naming + optional copy overrides (preferred over flat description/backLabel). */
  descriptions?: ResourceListAndEditDescriptions;
  /** Used with `descriptions` to pick create/edit/replicate intro copy. */
  mode?: ResourceEditMode;
  /** Optional intro copy above the field sections. Wins over `descriptions` when set. */
  description?: ReactNode;
  LinkComponent?: ComponentType<{
    href: string;
    className?: string;
    children: ReactNode;
  }>;
  renderRelatedLinkIcon?: (icon: string | undefined) => ReactNode;
  renderRelatedLinkChevron?: () => ReactNode;
  relatedLinksTitle?: string;
  onBack?: () => void;
  backLabel?: string;
};

export function ResourceEdit({
  state,
  handlers,
  saving = false,
  twoColumnLayout = true,
  locationPreviewDelayMs = DEFAULT_LOCATION_PREVIEW_DELAY_MS,
  descriptions,
  mode,
  description: descriptionProp,
  LinkComponent,
  renderRelatedLinkIcon,
  renderRelatedLinkChevron,
  relatedLinksTitle,
  onBack,
  backLabel: backLabelProp,
}: ResourceEditProps) {
  const copy = useMemo(() => resolveResourceListAndEditCopy(descriptions), [descriptions]);
  const description =
    descriptionProp ?? (mode ? resolveModeDescription(copy, mode) : undefined);
  const backLabel = backLabelProp ?? copy?.backLabel ?? "Back";
  const { onSubmit, onCreate, onUploadImage } = handlers;
  const [draft, setDraft] = useState(() => cloneResourceValues(state.values));
  const [baseline, setBaseline] = useState(() =>
    cloneResourceValues(state.baselineValues ?? state.values),
  );
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const previewTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const previewAbortRef = useRef<AbortController | null>(null);

  const sections = useMemo(
    () => groupFieldsBySection(state.schema.fields),
    [state.schema.fields],
  );

  const isDirty = useMemo(
    () => isResourceDraftDirty(state.schema, draft, baseline),
    [state.schema, draft, baseline],
  );

  const validationErrors = useMemo(() => {
    if (isResourceRecordNew(state.recordState)) {
      return validateResourceValues(state.schema, draft);
    }
    return validateDirtyFields(state.schema, draft, baseline);
  }, [state.schema, state.recordState, draft, baseline]);

  const isValid = firstValidationError(validationErrors) === null;
  const submitButton = getResourceSubmitButtonState({
    recordState: state.recordState,
    isDirty,
    saving: saving || submitting,
    isValid,
  });

  const syncFromServer = useCallback((values: Record<string, string>) => {
    const synced = cloneResourceValues(values);
    setBaseline(synced);
    setDraft(synced);
    setFieldErrors({});
  }, []);

  const clearLocationPreviewTimer = useCallback(() => {
    if (previewTimerRef.current) {
      clearTimeout(previewTimerRef.current);
      previewTimerRef.current = null;
    }
    previewAbortRef.current?.abort();
    previewAbortRef.current = null;
  }, []);

  useEffect(() => () => clearLocationPreviewTimer(), [clearLocationPreviewTimer]);

  const scheduleLocationPreview = useCallback(
    (nextDraft: Record<string, string>, changedKey: string) => {
      if (!fieldTriggersLocationPreview(state.schema.fields, changedKey)) {
        return;
      }
      const changed = state.schema.fields.find((field) => field.key === changedKey);
      const section = changed?.section || "general";
      const locationKeys = previewLocationFieldKeys(state.schema.fields);
      if (locationKeys.length === 0) {
        return;
      }

      // Don't overwrite a manual pin with address preview.
      const blocksPreview = locationKeys.some((key) => {
        const parsed = parseLocationValue(nextDraft[key] ?? "");
        return parsed?.source === "manual";
      });
      if (blocksPreview) {
        return;
      }

      clearLocationPreviewTimer();
      previewTimerRef.current = setTimeout(() => {
        previewTimerRef.current = null;
        const query = buildAddressQueryFromDraft(state.schema.fields, nextDraft, section);
        if (!query) {
          setDraft((current) => {
            const next = { ...current };
            for (const key of locationKeys) {
              next[key] = buildPendingLocationValue("");
            }
            return next;
          });
          return;
        }

        setDraft((current) => {
          const next = { ...current };
          for (const key of locationKeys) {
            next[key] = buildPendingLocationValue(query);
          }
          return next;
        });

        const controller = new AbortController();
        previewAbortRef.current = controller;
        void geocodeAddressPreview(query, controller.signal)
          .then((result) => {
            if (controller.signal.aborted) return;
            setDraft((current) => {
              const next = { ...current };
              for (const key of locationKeys) {
                next[key] = result
                  ? buildPreviewLocationValue(result.lat, result.lng, result.label ?? query)
                  : buildPendingLocationValue(query);
              }
              return next;
            });
          })
          .catch(() => {
            /* Preview is best-effort; Save still geocodes on the server. */
          });
      }, locationPreviewDelayMs);
    },
    [clearLocationPreviewTimer, locationPreviewDelayMs, state.schema.fields],
  );

  function handleChange(key: string, value: string) {
    const nextDraft = { ...draft, [key]: value };
    setDraft(nextDraft);
    scheduleLocationPreview(nextDraft, key);

    const keysToValidate = new Set([key]);
    const nextFieldErrors = validateResourceValues(state.schema, nextDraft, keysToValidate);
    setFieldErrors((current) => {
      const next = { ...current };
      for (const fieldKey of keysToValidate) {
        if (nextFieldErrors[fieldKey]) {
          next[fieldKey] = nextFieldErrors[fieldKey];
        } else {
          delete next[fieldKey];
        }
      }
      return next;
    });
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    clearLocationPreviewTimer();
    setError(null);

    if (!isDirty) {
      return;
    }

    const submitValues = normalizeSubmitValues(state.schema, buildSubmitValues(state.schema, draft));
    const submitValidationErrors = validateResourceValues(state.schema, submitValues);
    setFieldErrors(submitValidationErrors);
    const validationMessage = firstValidationError(submitValidationErrors);
    if (validationMessage) {
      setError(validationMessage);
      return;
    }

    const persist = isResourceRecordNew(state.recordState) && onCreate ? onCreate : onSubmit;

    setSubmitting(true);
    try {
      const nextState = await persist(submitValues);
      syncFromServer(nextState?.values ?? submitValues);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to save";
      const mapped = mapSubmitErrorToFieldErrors(state.schema, message);
      if (Object.keys(mapped.fieldErrors).length > 0) {
        setFieldErrors((current) => ({ ...current, ...mapped.fieldErrors }));
      }
      setError(mapped.formError);
    } finally {
      setSubmitting(false);
    }
  }

  const fieldGridClassName = twoColumnLayout
    ? "appkit-resource-edit-form__grid appkit-resource-edit-form__grid--two-col"
    : "appkit-resource-edit-form__grid";

  function renderBackButton() {
    if (!onBack) {
      return null;
    }
    return (
      <button
        type="button"
        className="appkit-resource-edit-button appkit-resource-edit-button--ghost"
        onClick={onBack}
      >
        <span className="appkit-resource-edit-button__back-icon" aria-hidden>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </span>
        {backLabel}
      </button>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="appkit-resource-edit-form" noValidate>
      {error ? <div className="appkit-resource-edit-form__error">{error}</div> : null}

      {onBack ? (
        <div className="appkit-resource-edit-form__back-top">{renderBackButton()}</div>
      ) : null}

      {description ? (
        <div className="appkit-resource-edit-form__description">{description}</div>
      ) : null}

      {sections.map(({ section, fields }) => (
        <section key={section} className="appkit-resource-edit-form__section">
          <h2 className="appkit-resource-edit-form__section-title">{formatSectionTitle(section)}</h2>
          <div className={fieldGridClassName}>
            {fields.map((field) => (
              <div
                key={field.key}
                className={
                  twoColumnLayout &&
                  (field.kind === ResourceFieldKind.TEXTAREA || field.kind === ResourceFieldKind.LOCATION)
                    ? "appkit-resource-edit-form__field appkit-resource-edit-form__field--wide"
                    : "appkit-resource-edit-form__field"
                }
              >
                <ResourceFieldInput
                  field={field}
                  value={draft[field.key] ?? ""}
                  error={fieldErrors[field.key] ?? validationErrors[field.key]}
                  onChange={handleChange}
                  onUploadImage={onUploadImage}
                />
              </div>
            ))}
          </div>
        </section>
      ))}

      <div className="appkit-resource-edit-form__actions">
        {renderBackButton()}
        <button
          type="submit"
          disabled={submitButton.disabled}
          aria-busy={submitButton.loading}
          className={`appkit-resource-edit-button appkit-resource-edit-button--primary appkit-resource-edit-form__submit${
            submitButton.loading ? " appkit-resource-edit-button--loading" : ""
          }`}
        >
          {submitButton.loading ? (
            <span className="appkit-resource-edit-button__spinner" aria-hidden />
          ) : null}
          {submitButton.label}
        </button>
      </div>

      <ResourceRelatedLinks
        links={state.relatedLinks as ResourceRelatedLink[]}
        title={relatedLinksTitle}
        LinkComponent={LinkComponent}
        renderIcon={renderRelatedLinkIcon}
        renderChevron={renderRelatedLinkChevron}
      />
    </form>
  );
}
