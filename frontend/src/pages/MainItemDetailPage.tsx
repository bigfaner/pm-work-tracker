import { useState, useCallback, useEffect } from "react";
import { useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTeamStore } from "@/store/team";
import { getMainItemApi, updateMainItemApi } from "@/api/mainItems";
import { createSubItemApi, updateSubItemApi } from "@/api/subItems";
import { appendProgressApi } from "@/api/progress";
import { listMembersApi } from "@/api/teams";
import type { SubItem } from "@/types";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PermissionGuard } from "@/components/PermissionGuard";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import PriorityBadge from "@/components/shared/PriorityBadge";
import StatusTransitionDropdown from "@/components/shared/StatusTransitionDropdown";
import { MAIN_ITEM_STATUSES } from "@/lib/status";
import { useMemberName } from "@/hooks/useMemberName";
import EditMainItemDialog, {
  type EditMainItemFormState,
} from "./main-item-detail/EditMainItemDialog";
import CreateSubItemDialog, {
  type CreateSubItemFormState,
} from "./main-item-detail/CreateSubItemDialog";
import EditSubItemDialog, {
  type EditSubItemFormState,
} from "./main-item-detail/EditSubItemDialog";
import AppendProgressDialog, {
  type AppendProgressFormState,
} from "./main-item-detail/AppendProgressDialog";
import SubItemsTable from "./main-item-detail/SubItemsTable";
import ProgressSummaryCard from "./main-item-detail/ProgressSummaryCard";
import ItemInfoCard from "./main-item-detail/ItemInfoCard";
import { DecisionTimeline } from "./main-item-detail/DecisionTimeline";
import { DecisionFormDialog } from "@/components/decision-log/DecisionFormDialog";

// --- Main Component ---

export default function MainItemDetailPage() {
  const { mainItemId } = useParams<{ mainItemId: string }>();
  const teamId = useTeamStore((s) => s.currentTeamId);
  const qc = useQueryClient();
  const itemId = mainItemId!;
  // State
  const [expanded, setExpanded] = useState(false);
  const today = () => new Date().toISOString().slice(0, 10);

  const [editOpen, setEditOpen] = useState(false);
  const [createSubOpen, setCreateSubOpen] = useState(false);
  const [editForm, setEditForm] = useState<EditMainItemFormState>({
    title: "",
    priority: "",
    assigneeKey: "",
    expectedEndDate: "",
    description: "",
  });
  const [subForm, setSubForm] = useState<CreateSubItemFormState>({
    title: "",
    priority: "P2",
    assigneeKey: "",
    startDate: today(),
    expectedEndDate: "",
    description: "",
  });

  const [editSubOpen, setEditSubOpen] = useState(false);
  const [editSubTarget, setEditSubTarget] = useState<SubItem | null>(null);
  const [editSubForm, setEditSubForm] = useState<EditSubItemFormState>({
    title: "",
    priority: "",
    expectedEndDate: "",
    description: "",
  });

  const [appendProgressOpen, setAppendProgressOpen] = useState(false);
  const [appendProgressTarget, setAppendProgressTarget] =
    useState<SubItem | null>(null);
  const [appendProgressForm, setAppendProgressForm] =
    useState<AppendProgressFormState>({
      completion: "",
      achievement: "",
      blocker: "",
    });

  // Decision form dialog state
  const [decisionFormOpen, setDecisionFormOpen] = useState(false);
  const [decisionFormMode, setDecisionFormMode] = useState<"new" | "edit">(
    "new",
  );
  const [decisionEditBizKey, setDecisionEditBizKey] = useState("");

  // --- Data fetching ---

  const { data: item, isLoading } = useQuery({
    queryKey: ["mainItem", teamId, itemId],
    queryFn: () => getMainItemApi(teamId!, itemId),
    enabled: !!teamId && !!itemId,
  });

  const { data: members } = useQuery({
    queryKey: ["members", teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  });

  const memberName = useMemberName(members);

  // Populate edit form when data loads
  useEffect(() => {
    if (item) {
      setEditForm({
        title: item.title,
        priority: item.priority,
        assigneeKey: item.assigneeKey || "",
        expectedEndDate: item.expectedEndDate || "",
        description: item.itemDesc || "",
      });
    }
  }, [item]);

  // --- Mutations ---

  const updateMutation = useMutation({
    mutationFn: (req: {
      title?: string;
      priority?: string;
      assigneeKey?: string | null;
      expectedEndDate?: string | null;
      actualEndDate?: string | null;
      description?: string;
    }) => updateMainItemApi(teamId!, itemId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["mainItem", teamId, itemId] });
      setEditOpen(false);
    },
  });

  const createSubMutation = useMutation({
    mutationFn: (req: {
      title: string;
      priority: string;
      assigneeKey: string;
      startDate?: string;
      expectedEndDate?: string;
      description?: string;
    }) => createSubItemApi(teamId!, itemId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["mainItem", teamId, itemId] });
      setCreateSubOpen(false);
      setSubForm({
        title: "",
        priority: "P2",
        assigneeKey: "",
        startDate: today(),
        expectedEndDate: "",
        description: "",
      });
    },
  });

  const updateSubMutation = useMutation({
    mutationFn: ({
      subId,
      req,
    }: {
      subId: string;
      req: {
        title?: string;
        priority?: string;
        expectedEndDate?: string;
        description?: string;
      };
    }) => updateSubItemApi(teamId!, subId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["mainItem", teamId, itemId] });
      setEditSubOpen(false);
      setEditSubTarget(null);
    },
  });

  const appendProgressMutation = useMutation({
    mutationFn: ({
      subId,
      req,
    }: {
      subId: string;
      req: { completion: number; achievement?: string; blocker?: string };
    }) => appendProgressApi(teamId!, subId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["mainItem", teamId, itemId] });
      setAppendProgressOpen(false);
      setAppendProgressTarget(null);
      setAppendProgressForm({ completion: "", achievement: "", blocker: "" });
    },
  });

  // --- Handlers ---

  const handleEdit = useCallback(() => {
    if (!editForm.title.trim()) return;
    updateMutation.mutate({
      title: editForm.title.trim(),
      priority: editForm.priority,
      assigneeKey: editForm.assigneeKey || null,
      expectedEndDate: editForm.expectedEndDate || null,
      description: editForm.description,
    });
  }, [editForm, updateMutation]);

  const handleCreateSub = useCallback(() => {
    if (
      !subForm.title.trim() ||
      !subForm.priority ||
      !subForm.assigneeKey ||
      !subForm.startDate ||
      !subForm.expectedEndDate
    )
      return;
    createSubMutation.mutate({
      title: subForm.title.trim(),
      priority: subForm.priority,
      assigneeKey: subForm.assigneeKey || "",
      startDate: subForm.startDate,
      expectedEndDate: subForm.expectedEndDate,
      ...(subForm.description && { description: subForm.description }),
    });
  }, [subForm, createSubMutation]);

  const openEditSub = useCallback((sub: SubItem) => {
    setEditSubTarget(sub);
    setEditSubForm({
      title: sub.title,
      priority: sub.priority,
      expectedEndDate: sub.expectedEndDate || "",
      description: sub.itemDesc || "",
    });
    setEditSubOpen(true);
  }, []);

  const handleEditSub = useCallback(() => {
    if (!editSubTarget || !editSubForm.title.trim()) return;
    updateSubMutation.mutate({
      subId: editSubTarget.bizKey,
      req: {
        title: editSubForm.title.trim(),
        priority: editSubForm.priority,
        expectedEndDate: editSubForm.expectedEndDate || undefined,
        description: editSubForm.description,
      },
    });
  }, [editSubTarget, editSubForm, updateSubMutation]);

  const openAppendProgress = useCallback((sub: SubItem) => {
    setAppendProgressTarget(sub);
    setAppendProgressForm({
      completion: String(sub.completion),
      achievement: "",
      blocker: "",
    });
    setAppendProgressOpen(true);
  }, []);

  const handleAppendProgress = useCallback(() => {
    if (!appendProgressTarget || appendProgressForm.completion === "") return;
    appendProgressMutation.mutate({
      subId: appendProgressTarget.bizKey,
      req: {
        completion: Number(appendProgressForm.completion),
        ...(appendProgressForm.achievement && {
          achievement: appendProgressForm.achievement,
        }),
        ...(appendProgressForm.blocker && {
          blocker: appendProgressForm.blocker,
        }),
      },
    });
  }, [appendProgressTarget, appendProgressForm, appendProgressMutation]);

  const subItems: SubItem[] = item?.subItems || [];
  const completedCount = subItems.filter(
    (s) => s.itemStatus === "completed",
  ).length;
  const completion = item?.completion ?? 0;

  // --- Render ---

  if (!teamId) return <div className="p-6 text-tertiary">请先选择团队</div>;

  return (
    <div data-testid="main-item-detail-page">
      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : !item ? (
        <div className="py-8 text-center text-tertiary text-sm">事项不存在</div>
      ) : (
        <>
          {/* Breadcrumb */}
          <Breadcrumb className="mb-4">
            <BreadcrumbItem href="/items">事项清单</BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem isCurrent>{item.title}</BreadcrumbItem>
          </Breadcrumb>
          {/* Title Bar */}
          <div className="flex items-center gap-3 mb-6 flex-wrap">
            <Badge variant="default" className="font-mono">
              {item.code}
            </Badge>
            <h1 className="text-xl font-semibold text-primary m-0">
              {item.title}
            </h1>
            <PriorityBadge priority={item.priority} />
            <StatusTransitionDropdown
              currentStatus={item.itemStatus}
              itemType="main"
              teamId={teamId!}
              itemId={item.bizKey}
              onStatusChanged={() => {
                qc.invalidateQueries({
                  queryKey: ["mainItem", teamId, itemId],
                });
              }}
            />
            <div className="flex-1" />
            <PermissionGuard code="main_item:update">
              <Button
                variant="secondary"
                disabled={
                  !!MAIN_ITEM_STATUSES[
                    item.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                  ]?.terminal
                }
                onClick={() => setEditOpen(true)}
              >
                <svg
                  width="16"
                  height="16"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
                编辑
              </Button>
            </PermissionGuard>
          </div>
          {/* Info Grid */}
          <ItemInfoCard
            assigneeName={memberName(item.assigneeKey)}
            startDate={item.planStartDate}
            expectedEndDate={item.expectedEndDate}
            actualEndDate={item.actualEndDate}
            status={item.itemStatus}
            description={item.itemDesc}
          />
          {/* Progress & Summary Card */}
          <ProgressSummaryCard
            completion={completion}
            completedCount={completedCount}
            totalSubItems={subItems.length}
            expanded={expanded}
            onToggleExpanded={() => setExpanded(!expanded)}
            achievements={item?.achievements}
            blockers={item?.blockers}
          />
          {/* Decision Timeline */}
          <DecisionTimeline
            teamId={String(teamId)}
            mainItemId={itemId}
            mainStatus={item.itemStatus}
            onAdd={() => {
              setDecisionFormMode("new");
              setDecisionEditBizKey("");
              setDecisionFormOpen(true);
            }}
            onEdit={(bizKey: string) => {
              setDecisionFormMode("edit");
              setDecisionEditBizKey(bizKey);
              setDecisionFormOpen(true);
            }}
          />
          {/* Sub-items Table */}
          <SubItemsTable
            subItems={subItems}
            mainItemId={item.bizKey}
            mainStatus={item.itemStatus}
            teamId={teamId!}
            memberName={memberName}
            onStatusChanged={() => {
              qc.invalidateQueries({ queryKey: ["mainItem", teamId, itemId] });
            }}
            onEditSub={openEditSub}
            onAppendProgress={openAppendProgress}
            onCreateSub={() => setCreateSubOpen(true)}
          />
          {/* Dialogs */}
          <EditMainItemDialog
            open={editOpen}
            onOpenChange={setEditOpen}
            form={editForm}
            onFormChange={setEditForm}
            members={members || []}
            onSubmit={handleEdit}
            isPending={updateMutation.isPending}
          />
          <CreateSubItemDialog
            open={createSubOpen}
            onOpenChange={setCreateSubOpen}
            form={subForm}
            onFormChange={setSubForm}
            members={members || []}
            onSubmit={handleCreateSub}
            isPending={createSubMutation.isPending}
          />
          <EditSubItemDialog
            open={editSubOpen}
            onOpenChange={setEditSubOpen}
            form={editSubForm}
            onFormChange={setEditSubForm}
            onSubmit={handleEditSub}
            isPending={updateSubMutation.isPending}
          />
          <AppendProgressDialog
            open={appendProgressOpen}
            onOpenChange={setAppendProgressOpen}
            form={appendProgressForm}
            onFormChange={setAppendProgressForm}
            onSubmit={handleAppendProgress}
            isPending={appendProgressMutation.isPending}
          />
          <DecisionFormDialog
            key={decisionEditBizKey}
            open={decisionFormOpen}
            onOpenChange={setDecisionFormOpen}
            mode={decisionFormMode}
            teamBizKey={String(teamId)}
            mainBizKey={item.bizKey}
            onSuccess={() => {
              qc.invalidateQueries({ queryKey: ["mainItem", teamId, itemId] });
              setDecisionFormOpen(false);
            }}
          />
        </>
      )}
    </div>
  );
}
