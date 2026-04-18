import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Table,
  Tag,
  Avatar,
  Button,
  Descriptions,
  Dropdown,
  Modal,
  Form,
  Input,
  Skeleton,
  Space,
  message,
} from 'antd'
import { UserAddOutlined, EllipsisOutlined, ExclamationCircleOutlined } from '@ant-design/icons'
import type { MenuProps } from 'antd'
import dayjs from 'dayjs'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import {
  getTeamApi,
  listMembersApi,
  inviteMemberApi,
  removeMemberApi,
  transferPmApi,
  deleteTeamApi,
} from '@/api/teams'
import type { TeamMemberResp } from '@/types'

export default function TeamManagementPage() {
  const { teamId } = useParams<{ teamId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const parsedTeamId = Number(teamId)

  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [dissolveModalOpen, setDissolveModalOpen] = useState(false)
  const [inviteForm] = Form.useForm()
  const [dissolveForm] = Form.useForm()
  const [inviteError, setInviteError] = useState<string | null>(null)
  const [dissolveName, setDissolveName] = useState('')

  // Queries
  const { data: team, isLoading: teamLoading } = useQuery({
    queryKey: ['team', parsedTeamId],
    queryFn: () => getTeamApi(parsedTeamId),
    enabled: !!parsedTeamId,
  })

  const { data: members = [], isLoading: membersLoading } = useQuery({
    queryKey: ['team-members', parsedTeamId],
    queryFn: () => listMembersApi(parsedTeamId),
    enabled: !!parsedTeamId,
  })

  const isPm = user?.id === team?.pmId

  // Mutations
  const inviteMutation = useMutation({
    mutationFn: (values: { username: string }) =>
      inviteMemberApi(parsedTeamId, { username: values.username, role: 'member' }),
    onSuccess: () => {
      message.success('邀请已发送')
      setInviteModalOpen(false)
      inviteForm.resetFields()
      setInviteError(null)
      queryClient.invalidateQueries({ queryKey: ['team-members', parsedTeamId] })
    },
    onError: (error: any) => {
      const code = error?.response?.data?.code
      if (code === 'USER_NOT_FOUND' || code === 'ALREADY_MEMBER') {
        setInviteError(error.response.data.message)
      }
    },
  })

  const removeMutation = useMutation({
    mutationFn: (userId: number) => removeMemberApi(parsedTeamId, userId),
    onSuccess: () => {
      message.success('已移除')
      queryClient.invalidateQueries({ queryKey: ['team-members', parsedTeamId] })
    },
  })

  const transferMutation = useMutation({
    mutationFn: (newPmUserId: number) => transferPmApi(parsedTeamId, { newPmUserId }),
    onSuccess: () => {
      message.success('已转让 PM 身份')
      queryClient.invalidateQueries({ queryKey: ['team', parsedTeamId] })
      queryClient.invalidateQueries({ queryKey: ['team-members', parsedTeamId] })
    },
  })

  const dissolveMutation = useMutation({
    mutationFn: (confirmName: string) => deleteTeamApi(parsedTeamId, { confirmName }),
    onSuccess: () => {
      message.success('团队已解散')
      navigate('/')
    },
  })

  // Loading state
  if (teamLoading || membersLoading) {
    return (
      <div data-testid="team-mgmt-skeleton" style={{ padding: 24 }}>
        <Skeleton active paragraph={{ rows: 8 }} />
      </div>
    )
  }

  // Action handlers
  function handleInvite() {
    setInviteModalOpen(true)
    setInviteError(null)
  }

  function handleInviteSubmit() {
    inviteForm.validateFields().then((values) => {
      setInviteError(null)
      inviteMutation.mutate(values)
    })
  }

  function handleRemove(member: TeamMemberResp) {
    Modal.confirm({
      icon: <ExclamationCircleOutlined />,
      content: `确认移除 ${member.displayName}？`,
      onOk: () => removeMutation.mutate(member.userId),
    })
  }

  function handleTransferPm(member: TeamMemberResp) {
    Modal.confirm({
      icon: <ExclamationCircleOutlined />,
      content: `确认将 ${member.displayName} 设为 PM？此操作不可撤销，您将降为普通成员。`,
      onOk: () => transferMutation.mutate(member.userId),
    })
  }

  function handleDissolve() {
    setDissolveModalOpen(true)
    setDissolveName('')
  }

  function handleDissolveConfirm() {
    dissolveForm.validateFields().then((values) => {
      dissolveMutation.mutate(values.confirmName)
    })
  }

  // Column definitions
  const columns = [
    {
      title: '姓名',
      dataIndex: 'displayName',
      key: 'displayName',
      render: (name: string) => (
        <Space>
          <Avatar size={24}>{name[0]}</Avatar>
          <span>{name}</span>
        </Space>
      ),
    },
    {
      title: '账号',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={role === 'pm' ? 'blue' : 'default'}>
          {role === 'pm' ? 'PM' : '成员'}
        </Tag>
      ),
    },
    {
      title: '加入时间',
      dataIndex: 'joinedAt',
      key: 'joinedAt',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: TeamMemberResp) => {
        // Hide actions for self (PM row)
        if (record.userId === user?.id) return null
        // Only PM can see actions
        if (!isPm) return null

        const menuItems: MenuProps['items'] = [
          ...(record.role !== 'pm'
            ? [{ key: 'transfer', label: '设为 PM', onClick: () => handleTransferPm(record) }]
            : []),
          { key: 'remove', label: '移除成员', danger: true, onClick: () => handleRemove(record) },
        ]

        return (
          <Dropdown menu={{ items: menuItems }} trigger={['click']}>
            <Button
              type="text"
              icon={<EllipsisOutlined />}
              data-testid={`member-actions-${record.userId}`}
            />
          </Dropdown>
        )
      },
    },
  ]

  return (
    <div data-testid="team-management-page" style={{ padding: 24, maxWidth: 960 }}>
      {/* Page Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2>团队管理</h2>
        {isPm && (
          <Button
            type="primary"
            icon={<UserAddOutlined />}
            onClick={handleInvite}
            data-testid="invite-member-btn"
          >
            邀请成员
          </Button>
        )}
      </div>

      {/* Team Info Card */}
      <Card style={{ marginBottom: 16 }} data-testid="team-info-card">
        <Descriptions column={3}>
          <Descriptions.Item label="团队名称">{team?.name}</Descriptions.Item>
          <Descriptions.Item label="PM">{team?.pm?.displayName}</Descriptions.Item>
          <Descriptions.Item label="成员数">{team?.memberCount}</Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {team?.createdAt ? dayjs(team.createdAt).format('YYYY-MM-DD') : '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Member Table */}
      <Table
        dataSource={members}
        columns={columns}
        rowKey="userId"
        size="middle"
        pagination={false}
        data-testid="member-table"
      />

      {/* Danger Zone - PM only */}
      {isPm && (
        <Card
          title="危险操作"
          style={{ marginTop: 32, borderColor: '#ff4d4f' }}
          data-testid="danger-zone"
        >
          <Button danger ghost onClick={handleDissolve} data-testid="dissolve-team-btn">
            解散团队
          </Button>
        </Card>
      )}

      {/* Invite Modal */}
      <Modal
        title="邀请成员"
        open={inviteModalOpen}
        onCancel={() => {
          setInviteModalOpen(false)
          inviteForm.resetFields()
          setInviteError(null)
        }}
        footer={[
          <Button key="cancel" onClick={() => { setInviteModalOpen(false); inviteForm.resetFields(); setInviteError(null) }}>
            取消
          </Button>,
          <Button key="submit" type="primary" loading={inviteMutation.isPending} onClick={handleInviteSubmit} data-testid="invite-submit-btn">
            邀请
          </Button>,
        ]}
        data-testid="invite-modal"
      >
        <Form form={inviteForm} layout="vertical">
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" data-testid="invite-username" />
          </Form.Item>
          {inviteError && (
            <div style={{ color: '#ff4d4f', marginBottom: 16 }}>{inviteError}</div>
          )}
        </Form>
      </Modal>

      {/* Dissolve Modal */}
      <Modal
        title="解散团队"
        open={dissolveModalOpen}
        onCancel={() => {
          setDissolveModalOpen(false)
          dissolveForm.resetFields()
          setDissolveName('')
        }}
        footer={null}
        data-testid="dissolve-modal"
      >
        <Form form={dissolveForm} layout="vertical">
          <p>请输入团队名称 <strong>{team?.name}</strong> 以确认解散：</p>
          <Form.Item
            name="confirmName"
            rules={[{ required: true, message: '请输入团队名称' }]}
          >
            <Input
              placeholder="请输入团队名称"
              data-testid="dissolve-name-input"
              onChange={(e) => setDissolveName(e.target.value)}
            />
          </Form.Item>
          <Button
            danger
            type="primary"
            disabled={dissolveName !== team?.name}
            loading={dissolveMutation.isPending}
            onClick={handleDissolveConfirm}
            data-testid="dissolve-confirm-btn"
          >
            确认解散
          </Button>
        </Form>
      </Modal>
    </div>
  )
}
