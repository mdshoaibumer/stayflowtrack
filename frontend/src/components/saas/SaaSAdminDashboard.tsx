'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { apiClient } from '@/lib/api';

interface SaaSMetrics {
  total_tenants: number;
  active_tenants: number;
  trialing_tenants: number;
  churned_tenants: number;
  mrr: number;
  arr: number;
  total_revenue: number;
  avg_revenue_per_tenant: number;
}

interface TenantOverview {
  tenant_id: string;
  tenant_name: string;
  owner_email: string;
  plan_name: string;
  plan_slug: string;
  status: string;
  billing_cycle: string;
  current_period_end: string;
  properties_count: number;
  units_count: number;
  users_count: number;
  created_at: string;
}

interface SubscriptionPlan {
  id: string;
  name: string;
  slug: string;
  price_monthly: number;
  price_yearly: number;
  max_properties: number;
  max_units: number;
  max_users: number;
  features: string[];
  is_active: boolean;
}

export default function SaaSAdminDashboard() {
  const [metrics, setMetrics] = useState<SaaSMetrics | null>(null);
  const [tenants, setTenants] = useState<TenantOverview[]>([]);
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [statusFilter, setStatusFilter] = useState('all');
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [metricsRes, tenantsRes, plansRes] = await Promise.all([
        apiClient.get('/admin/metrics'),
        apiClient.get(`/admin/tenants${statusFilter !== 'all' ? `?status=${statusFilter}` : ''}`),
        apiClient.get('/admin/plans'),
      ]);
      setMetrics(metricsRes.data);
      setTenants(tenantsRes.data.data || []);
      setPlans(plansRes.data || []);
    } catch (err) {
      console.error('Failed to load SaaS admin data:', err);
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 60000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR' }).format(amount);

  const statusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
      active: 'default',
      trialing: 'secondary',
      cancelled: 'destructive',
      past_due: 'destructive',
      expired: 'outline',
    };
    return <Badge variant={variants[status] || 'outline'}>{status}</Badge>;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">SaaS Admin</h1>
        <Button onClick={fetchData} variant="outline" size="sm">
          Refresh
        </Button>
      </div>

      {/* KPI Cards */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Total Tenants</CardDescription>
              <CardTitle className="text-2xl">{metrics.total_tenants}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground">
                {metrics.active_tenants} active · {metrics.trialing_tenants} trialing
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>MRR</CardDescription>
              <CardTitle className="text-2xl">{formatCurrency(metrics.mrr)}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground">
                ARR: {formatCurrency(metrics.arr)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Total Revenue</CardDescription>
              <CardTitle className="text-2xl">{formatCurrency(metrics.total_revenue)}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground">
                Avg/tenant: {formatCurrency(metrics.avg_revenue_per_tenant)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Churn</CardDescription>
              <CardTitle className="text-2xl">{metrics.churned_tenants}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground">
                {metrics.total_tenants > 0
                  ? ((metrics.churned_tenants / metrics.total_tenants) * 100).toFixed(1)
                  : 0}
                % churn rate
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Plans Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Subscription Plans</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {plans.map((plan) => (
              <Card key={plan.id} className={!plan.is_active ? 'opacity-50' : ''}>
                <CardHeader className="pb-2">
                  <CardTitle className="text-lg">{plan.name}</CardTitle>
                  <CardDescription>
                    {plan.price_monthly > 0
                      ? `${formatCurrency(plan.price_monthly)}/mo`
                      : 'Free'}
                  </CardDescription>
                </CardHeader>
                <CardContent className="text-sm space-y-1">
                  <p>{plan.max_properties} properties</p>
                  <p>{plan.max_units} units</p>
                  <p>{plan.max_users} users</p>
                  {!plan.is_active && <Badge variant="outline">Inactive</Badge>}
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Tenants Table */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Tenants</CardTitle>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="All statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="trialing">Trialing</SelectItem>
                <SelectItem value="past_due">Past Due</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Tenant</TableHead>
                <TableHead>Plan</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Properties</TableHead>
                <TableHead>Units</TableHead>
                <TableHead>Period End</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tenants.map((t) => (
                <TableRow key={t.tenant_id}>
                  <TableCell>
                    <div>
                      <p className="font-medium">{t.tenant_name}</p>
                      <p className="text-xs text-muted-foreground">{t.owner_email}</p>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{t.plan_name}</Badge>
                  </TableCell>
                  <TableCell>{statusBadge(t.status)}</TableCell>
                  <TableCell>{t.properties_count}</TableCell>
                  <TableCell>{t.units_count}</TableCell>
                  <TableCell>
                    {t.current_period_end
                      ? new Date(t.current_period_end).toLocaleDateString()
                      : '-'}
                  </TableCell>
                </TableRow>
              ))}
              {tenants.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    No tenants found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
