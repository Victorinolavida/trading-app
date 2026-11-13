import { Badge, Card, CardTitle, DetailItem, Section } from "./ui/card";
import { Security } from "../features/securities/types";

const typeLabel: Record<string, string> = {
  M: "Bono M",
  S: "UDI Bond",
};

export function SecurityInfoCard({ security }: { security: Security }) {
  return (
    <Card className="w-full">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <CardTitle className="mb-1">
            {security.name ||
              `Mexico Government Bond (${security.security_type})`}
          </CardTitle>
          <p className="text-sm text-fg-muted">
            {typeLabel[security.security_type] ?? security.security_type} ·{" "}
            {security.currency}
          </p>
        </div>
        <Badge active={security.enable} />
      </div>

      <Section title="Identification">
        <DetailItem label="ISIN" mono>
          {security.isin}
        </DetailItem>
        <DetailItem label="Serie" mono>
          {security.serie}
        </DetailItem>
        <DetailItem label="ID" mono className="sm:col-span-2">
          {security.id}
        </DetailItem>
      </Section>

      <Section title="Terms">
        <DetailItem label="Nominal value">{security.nominal_value}</DetailItem>
        <DetailItem label="Currency">{security.currency}</DetailItem>
        <DetailItem label="FX">{security.fx}</DetailItem>
      </Section>

      <Section title="Dates">
        <DetailItem label="Issue date">{security.issue_date}</DetailItem>
        <DetailItem label="Maturity date">{security.maturity_date}</DetailItem>
      </Section>
    </Card>
  );
}
