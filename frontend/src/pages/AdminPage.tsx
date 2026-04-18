import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Tabs, Table, Switch, Tag, Skeleton, Tooltip, message } from 'antd'
import dayjs from 'dayjs'
import { useAuthStore } from '@/store/auth'
import { listUsersApi, setCanCreateTeamApi, listAdminTeamsApi } from '@/api/admin'
import type { AdminUser, AdminTeam } from '@/types'
import type { ColumnsType } from 'antd/es/table'

const PAGE_SIZE = 50

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState('users')
  const currentUser = useAuthStore((s) => s.user)
  const queryClient = useQueryClient()

  // --- Users query ---
  const [usersPage, setUsersPage] = useState(1)
  const {
    data: usersData,
    isLoading: usersLoading,
  } = useQuery({
    queryKey: ['admin-users', usersPage],
    queryFn: () => listUsersApi(usersPage, PAGE_SIZE),
  })

  // --- Teams query ---
  const [teamsPage, setTeamsPage] = useState(1)
  const {
    data: teamsData,
    isLoading: teamsLoading,
  } = useQuery({
    queryKey: ['admin-teams', teamsPage],
    queryFn: () => listAdminTeamsApi(teamsPage, PAGE_SIZE),
    enabled: activeTab === 'teams',
  })

  // --- Toggle mutation ---
  const toggleMutation = useMutation({
    mutationFn: ({ userId, canCreateTeam }: { userId: number; canCreateTeam: boolean }) =>
      setCanCreateTeamApi(userId, { canCreateTeam }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
    },
    onError: () => {
      message.error('更新失败')
    },
  })

  // --- User columns ---
  const userColumns: ColumnsType<AdminUser> = [
    {
      title: '账号',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '姓名',
      dataIndex: 'displayName',
      key: 'displayName',
    },
    {
      title: '创建团队权限',
      dataIndex: 'canCreateTeam',
      key: 'canCreateTeam',
      render: (val: boolean, record: AdminUser) => {
        const isSelf = record.id === currentUser?.id
        const isLoading = toggleMutation.isPending && toggleMutation.variables?.userId === record.id

        const switchEl = (
          <Switch
            checked={val}
            loading={isLoading}
            disabled={isSelf}
            onChange={(checked) => {
              toggleMutation.mutate({ userId: record.id, canCreateTeam: checked })
            }}
          />
        )

        if (isSelf) {
          return <Tooltip title="不能修改自己的权限">{switchEl}</Tooltip>
        }
        return switchEl
      },
    },
    {
      title: '超级管理员',
      dataIndex: 'isSuperAdmin',
      key: 'isSuperAdmin',
      render: (val: boolean) =>
        val ? <Tag color="blue">超级管理员</Tag> : null,
    },
  ]

  // --- Team columns ---
  const teamColumns: ColumnsType<AdminTeam> = [
    {
      title: '团队名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'PM姓名',
      dataIndex: 'pm',
      key: 'pm',
      render: (pm: { displayName: string }) => pm.displayName,
    },
    {
      title: '成员数',
      dataIndex: 'memberCount',
      key: 'memberCount',
    },
    {
      title: '主事项数',
      dataIndex: 'mainItemCount',
      key: 'mainItemCount',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
    },
  ]

  const tabItems = [
    {
      key: 'users',
      label: '用户管理',
      children: usersLoading ? (
        <div data-testid="admin-users-skeleton">
          <Skeleton active paragraph={{ rows: 6 }} />
        </div>
      ) : (
        <div data-testid="admin-users-table">
          <Table<AdminUser>
            dataSource={usersData?.items}
            columns={userColumns}
            rowKey="id"
            size="middle"
            pagination={{
              current: usersPage,
              pageSize: PAGE_SIZE,
              total: usersData?.total ?? 0,
              onChange: (page) => setUsersPage(page),
            }}
          />
        </div>
      ),
    },
    {
      key: 'teams',
      label: '团队列表',
      children: teamsLoading ? (
        <div data-testid="admin-teams-skeleton">
          <Skeleton active paragraph={{ rows: 6 }} />
        </div>
      ) : (
        <div data-testid="admin-teams-table">
          <Table<AdminTeam>
            dataSource={teamsData?.items}
            columns={teamColumns}
            rowKey="id"
            size="middle"
            pagination={{
              current: teamsPage,
              pageSize: PAGE_SIZE,
              total: teamsData?.total ?? 0,
              onChange: (page) => setTeamsPage(page),
            }}
          />
        </div>
      ),
    },
  ]

  return (
    <div data-testid="admin-page" style={{ padding: 24 }}>
      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key)}
        items={tabItems}
      />
    </div>
  )
}
