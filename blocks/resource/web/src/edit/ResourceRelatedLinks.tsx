"use client";

import type { ComponentType, ReactNode } from "react";
import type { ResourceRelatedLink } from "./types";

export type ResourceRelatedLinksProps = {
  links: ResourceRelatedLink[];
  title?: string;
  LinkComponent?: ComponentType<{
    href: string;
    className?: string;
    children: ReactNode;
  }>;
  renderIcon?: (icon: string | undefined) => ReactNode;
  renderChevron?: () => ReactNode;
};

function DefaultLink({
  href,
  className,
  children,
}: {
  href: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <a href={href} className={className}>
      {children}
    </a>
  );
}

export function ResourceRelatedLinks({
  links,
  title = "Related settings",
  LinkComponent = DefaultLink,
  renderIcon,
  renderChevron,
}: ResourceRelatedLinksProps) {
  if (links.length === 0) {
    return null;
  }

  return (
    <section className="appkit-resource-edit-related">
      <h2 className="appkit-resource-edit-related__title">{title}</h2>
      <ul className="appkit-resource-edit-related__list">
        {links.map((link) => (
          <li key={link.route}>
            <LinkComponent
              href={link.route}
              className="appkit-resource-edit-related__link"
            >
              <span className="appkit-resource-edit-related__icon">
                {renderIcon?.(link.icon)}
              </span>
              <span className="appkit-resource-edit-related__body">
                <span className="appkit-resource-edit-related__label">{link.label}</span>
                {link.description ? (
                  <span className="appkit-resource-edit-related__description">{link.description}</span>
                ) : null}
              </span>
              {renderChevron ? (
                <span className="appkit-resource-edit-related__chevron">{renderChevron()}</span>
              ) : null}
            </LinkComponent>
          </li>
        ))}
      </ul>
    </section>
  );
}
