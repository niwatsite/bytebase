import axios from "axios";
import {
  Member,
  MemberId,
  MemberCreate,
  MemberPatch,
  MemberState,
  ResourceObject,
  PrincipalId,
  unknown,
  empty,
  EMPTY_ID,
  Principal,
  ResourceIdentifier,
} from "../../types";

function convert(
  member: ResourceObject,
  includedList: ResourceObject[],
  rootGetters: any
): Member {
  const creatorId = (member.relationships!.creator.data as ResourceIdentifier)
    .id;
  let creator: Principal = unknown("PRINCIPAL") as Principal;
  creator.id = creatorId;

  const updaterId = (member.relationships!.updater.data as ResourceIdentifier)
    .id;
  let updater: Principal = unknown("PRINCIPAL") as Principal;
  updater.id = updaterId;

  const principalId = (member.relationships!.updater.data as ResourceIdentifier)
    .id;
  let principal: Principal = unknown("PRINCIPAL") as Principal;
  principal.id = principalId;

  for (const item of includedList || []) {
    if (
      item.type == "principal" &&
      (member.relationships!.creator.data as ResourceIdentifier).id == item.id
    ) {
      creator = rootGetters["principal/convert"](item);
    }

    if (
      item.type == "principal" &&
      (member.relationships!.updater.data as ResourceIdentifier).id == item.id
    ) {
      updater = rootGetters["principal/convert"](item);
    }

    if (
      item.type == "principal" &&
      (member.relationships!.principal.data as ResourceIdentifier).id == item.id
    ) {
      principal = rootGetters["principal/convert"](item);
    }
  }

  return {
    ...(member.attributes as Omit<
      Member,
      "id" | "creator" | "updater" | "principal"
    >),
    id: member.id,
    creator,
    updater,
    principal,
  };
}

const state: () => MemberState = () => ({
  memberList: [],
});

const getters = {
  memberList: (state: MemberState) => (): Member[] => {
    return state.memberList;
  },
  memberByPrincipalId:
    (state: MemberState) =>
    (id: PrincipalId): Member => {
      if (id == EMPTY_ID) {
        return empty("MEMBER") as Member;
      }

      return (
        state.memberList.find((item) => item.principal.id == id) ||
        (unknown("MEMBER") as Member)
      );
    },
};

const actions = {
  async fetchMemberList({ commit, rootGetters }: any) {
    const data = (await axios.get(`/api/member`)).data;
    const memberList = data.data.map((member: ResourceObject) => {
      return convert(member, data.included, rootGetters);
    });

    commit("setMemberList", memberList);
    return memberList;
  },

  // Returns existing member if the principalId has already been created.
  async createdMember({ commit, rootGetters }: any, newMember: MemberCreate) {
    const data = (
      await axios.post(`/api/member`, {
        data: {
          type: "MemberCreate",
          attributes: newMember,
        },
      })
    ).data;
    const createdMember = convert(data.data, data.included, rootGetters);

    commit("appendMember", createdMember);

    return createdMember;
  },

  async patchMember(
    { commit, rootGetters }: any,
    { id, memberPatch }: { id: MemberId; memberPatch: MemberPatch }
  ) {
    const data = (
      await axios.patch(`/api/member/${id}`, {
        data: {
          type: "memberpatch",
          attributes: memberPatch,
        },
      })
    ).data;
    const updatedMember = convert(data.data, data.included, rootGetters);

    commit("replaceMemberInList", updatedMember);

    return updatedMember;
  },

  async deleteMemberById(
    { state, commit }: { state: MemberState; commit: any },
    id: MemberId
  ) {
    await axios.delete(`/api/member/${id}`);

    const newList = state.memberList.filter((item: Member) => {
      return item.id != id;
    });

    commit("setMemberList", newList);
  },
};

const mutations = {
  setMemberList(state: MemberState, memberList: Member[]) {
    state.memberList = memberList;
  },

  appendMember(state: MemberState, newMember: Member) {
    state.memberList.push(newMember);
  },

  replaceMemberInList(state: MemberState, updatedMember: Member) {
    const i = state.memberList.findIndex(
      (item: Member) => item.id == updatedMember.id
    );
    if (i != -1) {
      state.memberList[i] = updatedMember;
    }
  },
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations,
};
