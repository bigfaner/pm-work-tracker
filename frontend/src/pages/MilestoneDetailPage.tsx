import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import {
  updateMilestoneMapApi,
  updateMilestoneApi,
} from '@/api/milestones'
import { listMembersApi } from '@/api/teams'
import MilestoneTimeline from './milestones/MilestoneTimeline'
import CreateMilestoneMapDialog, {
  type MilestoneMapFormState,
} from './milestones/CreateMilestoneMapDialog'
import CreateMilestoneDialog, {
  type MilestoneFormState,
} from './milestones/CreateMilestoneDialog'
import type { MilestoneMap, Milestone } from '@/types'

const EMPTY_MAP_FORM: MilestoneMapFormState = {
  mapName: '',
  assigneeKey: '',
  planStartDate: '',
  expectedEndDate: '',
  mapDesc: '',
}

const EMPTY_MILESTONE_FORM: MilestoneFormState = {
  milestoneName: '',
  expectedEndDate: '',
  milestoneDesc: '',
}

export default function MilestoneDetailPage() {
  const { mapId } = useParams<{ mapId: string }>()
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()

  // Edit map dialog
  const [editMapOpen, setEditMapOpen] = useState(false)
  const [editMapTarget, setEditMapTarget] = useState<MilestoneMap | null>(null)
  const [editMapForm, setEditMapForm] =
    useState<MilestoneMapFormState>(EMPTY_MAP_FORM)

  // Edit milestone dialog
  const [editMilestoneOpen, setEditMilestoneOpen] = useState(false)
  const [editMilestoneTarget, setEditMilestoneTarget] =
    useState<Milestone | null>(null)
  const [editMilestoneForm, setEditMilestoneForm] =
    useState<MilestoneFormState>(EMPTY_MILESTONE_FORM)

  // Members for map edit dialog
  const { data: membersData } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const members = (membersData || []).map(
    (m: { userKey: string, displayName: string }) => ({
      userKey: m.userKey,
      displayName: m.displayName,
    }),
  )

  // Populate map edit form when target changes
  useEffect(() => {
    if (editMapTarget) {
      setEditMapForm({
        mapName: editMapTarget.mapName,
        assigneeKey: editMapTarget.assigneeKey,
        planStartDate: editMapTarget.planStartDate ?? '',
        expectedEndDate: editMapTarget.expectedEndDate ?? '',
        mapDesc: editMapTarget.mapDesc ?? '',
      })
    }
  }, [editMapTarget])

  // Populate milestone edit form when target changes
  useEffect(() => {
    if (editMilestoneTarget) {
      setEditMilestoneForm({
        milestoneName: editMilestoneTarget.milestoneName,
        expectedEndDate: editMilestoneTarget.expectedEndDate ?? '',
        milestoneDesc: editMilestoneTarget.milestoneDesc ?? '',
      })
    }
  }, [editMilestoneTarget])

  // Update map mutation
  const updateMapMutation = useMutation({
    mutationFn: (form: MilestoneMapFormState) =>
      updateMilestoneMapApi(teamId!, mapId!, {
        mapName: form.mapName.trim(),
        assigneeBizKey: form.assigneeKey,
        ...(form.planStartDate && { planStartDate: form.planStartDate }),
        ...(form.expectedEndDate && { expectedEndDate: form.expectedEndDate }),
        ...(form.mapDesc && { mapDesc: form.mapDesc }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestoneMap', teamId, mapId] })
      qc.invalidateQueries({ queryKey: ['milestoneMaps', teamId] })
      setEditMapOpen(false)
      setEditMapTarget(null)
    },
  })

  // Update milestone mutation
  const updateMilestoneMutation = useMutation({
    mutationFn: (form: MilestoneFormState) =>
      updateMilestoneApi(teamId!, editMilestoneTarget!.bizKey, {
        milestoneName: form.milestoneName.trim(),
        ...(form.expectedEndDate && { expectedEndDate: form.expectedEndDate }),
        ...(form.milestoneDesc && { milestoneDesc: form.milestoneDesc }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones', teamId, mapId] })
      qc.invalidateQueries({ queryKey: ['milestone', teamId] })
      setEditMilestoneOpen(false)
      setEditMilestoneTarget(null)
    },
  })

  // Callbacks for MilestoneTimeline
  const handleEditMap = (map: MilestoneMap) => {
    setEditMapTarget(map)
    setEditMapOpen(true)
  }

  const handleEditMilestone = (milestone: Milestone) => {
    setEditMilestoneTarget(milestone)
    setEditMilestoneOpen(true)
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const handleQuickAdd = (_milestone: Milestone) => {
    // UF-3a quick add dialog — placeholder until implemented
  }

  return (
    <>
      <MilestoneTimeline
        mapId={mapId!}
        onEditMap={handleEditMap}
        onEditMilestone={handleEditMilestone}
        onQuickAdd={handleQuickAdd}
      />

      {/* Edit map dialog */}
      <CreateMilestoneMapDialog
        open={editMapOpen}
        onOpenChange={(open) => {
          setEditMapOpen(open)
          if (!open) setEditMapTarget(null)
        }}
        form={editMapForm}
        onFormChange={setEditMapForm}
        members={members}
        onSubmit={() => updateMapMutation.mutate(editMapForm)}
        isPending={updateMapMutation.isPending}
        milestoneMap={editMapTarget ?? undefined}
      />

      {/* Edit milestone dialog */}
      <CreateMilestoneDialog
        open={editMilestoneOpen}
        onOpenChange={(open) => {
          setEditMilestoneOpen(open)
          if (!open) setEditMilestoneTarget(null)
        }}
        form={editMilestoneForm}
        onFormChange={setEditMilestoneForm}
        onSubmit={() => updateMilestoneMutation.mutate(editMilestoneForm)}
        isPending={updateMilestoneMutation.isPending}
        milestone={editMilestoneTarget ?? undefined}
      />
    </>
  )
}
