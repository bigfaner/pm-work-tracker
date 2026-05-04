import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Member {
  userKey: string;
  displayName: string;
}

interface MemberSelectProps {
  members: Member[];
  selectedId: string;
  onSelect: (id: string) => void;
  placeholder?: string;
  allowEmpty?: boolean;
}

export function MemberSelect({
  members,
  selectedId,
  onSelect,
  placeholder = "选择负责人",
  allowEmpty = true,
}: MemberSelectProps) {
  const value = selectedId || (allowEmpty ? "_none" : undefined);

  return (
    <Select
      value={value}
      onValueChange={(v) => onSelect(v === "_none" ? "" : v)}
    >
      <SelectTrigger>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {allowEmpty && <SelectItem value="_none">不指定</SelectItem>}
        {members.map((m) => (
          <SelectItem key={m.userKey} value={m.userKey}>
            {m.displayName}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
