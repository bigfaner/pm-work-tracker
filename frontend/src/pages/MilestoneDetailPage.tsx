import { useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import MilestoneTimeline from './milestones/MilestoneTimeline'
import CreateMilestoneDialog, {
  type MilestoneFormState,
} from './milestones/CreateMilestoneDialog'
import CreateMilestoneMapDialog, {
  type MilestoneMapFormState,
} from './milestones/CreateMilestoneMapDialog'
import type { Milestone, MilestoneMap } from '@/types'

export default function MilestoneDetailPage() {
  const { mapId } = useParams<{ mapId: string }>()

  // Edit milestone dialog state
  const [editMilestoneDialogOpen, setEditMilestoneDialogOpen] = useState(false)
  const [editMilestone, setEditMilestone] = useState<Milestone | null>(null)
  const [editMilestoneForm, setEditMilestoneForm] =
    useState<MilestoneFormState>({
      milestoneName: '',
      expectedEndDate: '',
      milestoneDesc: '',
    })

  // Edit map dialog state
  const [editMapDialogOpen, setEditMapDialogOpen] = useState(false)
  const [editMap, setEditMap] = useState<MilestoneMap | null>(null)
  const [editMapForm, setEditMapForm] = useState<MilestoneMapFormState>({
    mapName: '',
    assigneeKey: '',
    planStartDate: '',
    expectedEndDate: '',
    mapDesc: '',
  })

  const handleEditMilestone = useCallback((milestone: Milestone) => {
    setEditMilestone(milestone)
    setEditMilestoneForm({
      milestoneName: milestone.milestoneName,
      expectedEndDate: milestone.expectedEndDate ?? '',
      milestoneDesc: milestone.milestoneDesc ?? '',
    })
    setEditMilestoneDialogOpen(true)
  }, [])

  const handleEditMap = useCallback((map: MilestoneMap) => {
    setEditMap(map)
    setEditMapForm({
      mapName: map.mapName,
      assigneeKey: map.assigneeKey,
      planStartDate: map.planStartDate ?? '',
      expectedEndDate: map.expectedEndDate ?? '',
      mapDesc: map.mapDesc ?? '',
    })
    setEditMapDialogOpen(true)
  }, [])

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const handleQuickAdd = useCallback((_m: Milestone) => {
    // Will be connected to QuickAddMainItemDialog in a future task
  }, [])

  if (!mapId) return null

  return (
    <>
      <MilestoneTimeline
        mapId={mapId}
        onEditMap={handleEditMap}
        onEditMilestone={handleEditMilestone}
        onQuickAdd={handleQuickAdd}
      />
      {/* Edit milestone dialog placeholder - will be fully wired in integration */}
      <CreateMilestoneDialog
        open={editMilestoneDialogOpen}
        onOpenChange={setEditMilestoneDialogOpen}
        form={editMilestoneForm}
        onFormChange={setEditMilestoneForm}
        onSubmit={() => setEditMilestoneDialogOpen(false)}
        isPending={false}
        milestone={editMilestone ?? undefined}
      />
      {/* Edit map dialog placeholder */}
      <CreateMilestoneMapDialog
        open={editMapDialogOpen}
        onOpenChange={setEditMapDialogOpen}
        form={editMapForm}
        onFormChange={setEditMapForm}
        members={[]}
        onSubmit={() => setEditMapDialogOpen(false)}
        isPending={false}
        milestoneMap={editMap ?? undefined}
      />
    </>
  )
}
