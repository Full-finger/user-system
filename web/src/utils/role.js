const ROLE_LABELS = {
  user: '普通用户',
  verified_user: '认证用户',
  moderator: '版主',
  admin: '管理员',
  super_admin: '超级管理员',
}

export function roleLabel(role) {
  return ROLE_LABELS[role] || role
}

export const ADMIN_ROLES = ['admin', 'super_admin']
export const MANAGE_ROLES = ['moderator', ...ADMIN_ROLES]

export const ASSIGNABLE_ROLES = {
  admin: ['user', 'verified_user'],
  super_admin: ['user', 'verified_user', 'admin'],
}
