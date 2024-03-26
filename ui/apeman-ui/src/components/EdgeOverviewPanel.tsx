import { Relationship } from "../services/relationshipServices";

interface Props {
  edge: Relationship;
}
const EdgeOverviewPanel = ({ edge }: Props) => {
  return <div>{edge.ID}</div>;
};

export default EdgeOverviewPanel;
