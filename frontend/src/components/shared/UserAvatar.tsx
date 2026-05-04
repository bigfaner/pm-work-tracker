import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

const colors = [
  "bg-blue-100 text-blue-700",
  "bg-success-bg text-success-text",
  "bg-error-bg text-error-text",
  "bg-warning-bg text-warning-text",
  "bg-purple-100 text-purple-700",
  "bg-pink-100 text-pink-700",
  "bg-teal-100 text-teal-700",
];

function getColorForName(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
}

const sizeClasses = {
  sm: "w-7 h-7 text-[11px]",
  md: "w-8 h-8 text-[13px]",
  lg: "w-10 h-10 text-[15px]",
};

interface UserAvatarProps {
  name: string;
  src?: string;
  size?: "sm" | "md" | "lg";
  className?: string;
}

export default function UserAvatar({
  name,
  src,
  size = "md",
  className,
}: UserAvatarProps) {
  const initial = name ? name.charAt(0) : "?";
  const colorClass = getColorForName(name);

  return (
    <Avatar className={cn(sizeClasses[size], className)}>
      {src && (
        <img
          src={src}
          alt={name}
          className="aspect-square h-full w-full object-cover"
        />
      )}
      <AvatarFallback className={colorClass}>{initial}</AvatarFallback>
    </Avatar>
  );
}
