import * as React from "react";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";

function Breadcrumb({ className, ...props }: React.HTMLAttributes<HTMLElement>) {
  return (
    <nav
      aria-label="Breadcrumb"
      className={cn(
        "flex items-center gap-1.5 text-[13px] text-tertiary",
        className,
      )}
      {...props}
    />
  );
}

interface BreadcrumbItemProps extends React.LiHTMLAttributes<HTMLLIElement> {
  href?: string;
  isCurrent?: boolean;
}

function BreadcrumbItem({
  href,
  isCurrent,
  className,
  children,
  ...props
}: BreadcrumbItemProps) {
  return (
    <li
      className={cn("inline-flex items-center gap-1.5", className)}
      {...props}
    >
      {href && !isCurrent ? (
        <Link
          to={href}
          className="text-tertiary hover:text-primary-600 transition-colors"
        >
          {children}
        </Link>
      ) : (
        <span className={isCurrent ? "text-primary font-medium" : undefined}>
          {children}
        </span>
      )}
    </li>
  );
}

function BreadcrumbSeparator({
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn("text-tertiary", className)}
      aria-hidden="true"
      {...props}
    >
      /
    </span>
  );
}

export { Breadcrumb, BreadcrumbItem, BreadcrumbSeparator };
